package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	parkingdomain "apihorpug/internal/features/parking/domain"
	parkingusecase "apihorpug/internal/features/parking/usecase"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const selectParkingColumns = `
	p.id, p.tenant_id, t.first_name || ' ' || t.last_name, p.room_id, rm.room_number, rm.dormitory_id, d.name,
	p.vehicle_type, p.license_plate, p.parking_spot, p.created_by, p.updated_by, p.created_at, p.updated_at
`

const parkingFromJoins = `
	FROM parking_registrations p
	JOIN tenants t ON t.id = p.tenant_id
	LEFT JOIN rooms rm ON rm.id = p.room_id
	LEFT JOIN dormitories d ON d.id = rm.dormitory_id
`

// buildScope restricts results to dormitories the requester manages. Rows
// with no room assigned have no dormitory to check against, so they only
// ever match for roles with full dormitory access.
func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters parkingusecase.ListFilters, argIdx *int, args *[]any) []string {
	conditions := make([]string, 0)

	if !full {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, *argIdx, *argIdx+1))
		*args = append(*args, requesterID, roleID)
		*argIdx += 2
	}
	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf(`p.tenant_id = $%d`, *argIdx))
		*args = append(*args, *filters.TenantID)
		*argIdx++
	}
	if filters.RoomID != nil {
		conditions = append(conditions, fmt.Sprintf(`p.room_id = $%d`, *argIdx))
		*args = append(*args, *filters.RoomID)
		*argIdx++
	}
	if filters.DormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id = $%d`, *argIdx))
		*args = append(*args, *filters.DormitoryID)
		*argIdx++
	}
	if filters.VehicleType != nil {
		conditions = append(conditions, fmt.Sprintf(`p.vehicle_type = $%d`, *argIdx))
		*args = append(*args, *filters.VehicleType)
		*argIdx++
	}

	return conditions
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters parkingusecase.ListFilters) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT COUNT(*) ` + parkingFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters parkingusecase.ListFilters, limit, offset int) ([]parkingdomain.Parking, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT ` + selectParkingColumns + parkingFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY p.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanParkings(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (parkingdomain.Parking, error) {
	if err := r.ensureParkingAccess(ctx, id, requesterID); err != nil {
		return parkingdomain.Parking{}, err
	}

	parking, err := r.loadParkingByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return parkingdomain.Parking{}, parkingdomain.ErrParkingNotFound
		}
		return parkingdomain.Parking{}, err
	}
	return parking, nil
}

func (r *Repository) Create(ctx context.Context, input parkingusecase.CreateInput) (parkingdomain.Parking, error) {
	if input.RoomID != nil && input.CreatedBy != nil {
		if err := r.ensureRoomAccess(ctx, *input.RoomID, *input.CreatedBy); err != nil {
			return parkingdomain.Parking{}, err
		}
	}

	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		INSERT INTO parking_registrations (id, tenant_id, room_id, vehicle_type, license_plate, parking_spot, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`, id, input.TenantID, input.RoomID, input.VehicleType, input.LicensePlate, input.ParkingSpot, input.CreatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			if pgErr.ConstraintName == "parking_registrations_room_fkey" {
				return parkingdomain.Parking{}, parkingdomain.ErrRoomNotFound
			}
			return parkingdomain.Parking{}, parkingdomain.ErrTenantNotFound
		}
		return parkingdomain.Parking{}, err
	}

	return r.loadParkingByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input parkingusecase.UpdateInput) (parkingdomain.Parking, error) {
	if err := r.ensureParkingAccess(ctx, id, requesterID); err != nil {
		return parkingdomain.Parking{}, err
	}
	if input.RoomID != nil {
		if err := r.ensureRoomAccess(ctx, *input.RoomID, requesterID); err != nil {
			return parkingdomain.Parking{}, err
		}
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.RoomID != nil {
		setClauses = append(setClauses, fmt.Sprintf("room_id = $%d", argIdx))
		args = append(args, *input.RoomID)
		argIdx++
	}
	if input.VehicleType != nil {
		setClauses = append(setClauses, fmt.Sprintf("vehicle_type = $%d", argIdx))
		args = append(args, *input.VehicleType)
		argIdx++
	}
	if input.LicensePlate != nil {
		setClauses = append(setClauses, fmt.Sprintf("license_plate = $%d", argIdx))
		args = append(args, *input.LicensePlate)
		argIdx++
	}
	if input.ParkingSpot != nil {
		setClauses = append(setClauses, fmt.Sprintf("parking_spot = $%d", argIdx))
		args = append(args, *input.ParkingSpot)
		argIdx++
	}
	if input.UpdatedBy != nil {
		setClauses = append(setClauses, fmt.Sprintf("updated_by = $%d", argIdx))
		args = append(args, *input.UpdatedBy)
		argIdx++
	}

	if len(setClauses) > 0 {
		setClauses = append(setClauses, "updated_at = NOW()")
		args = append(args, id)
		query := fmt.Sprintf("UPDATE parking_registrations SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			return parkingdomain.Parking{}, err
		}
	}

	return r.loadParkingByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureParkingAccess(ctx, id, requesterID); err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, `DELETE FROM parking_registrations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return parkingdomain.ErrParkingNotFound
	}

	return nil
}

func (r *Repository) loadParkingByID(ctx context.Context, id uuid.UUID) (parkingdomain.Parking, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectParkingColumns+parkingFromJoins+` WHERE p.id = $1`, id)
	return scanParking(row)
}

func scanParking(row pgx.Row) (parkingdomain.Parking, error) {
	var parking parkingdomain.Parking
	if err := row.Scan(
		&parking.ID,
		&parking.TenantID,
		&parking.TenantName,
		&parking.RoomID,
		&parking.RoomNumber,
		&parking.DormitoryID,
		&parking.DormitoryName,
		&parking.VehicleType,
		&parking.LicensePlate,
		&parking.ParkingSpot,
		&parking.CreatedBy,
		&parking.UpdatedBy,
		&parking.CreatedAt,
		&parking.UpdatedAt,
	); err != nil {
		return parkingdomain.Parking{}, err
	}
	return parking, nil
}

func scanParkings(rows pgx.Rows) ([]parkingdomain.Parking, error) {
	parkings := make([]parkingdomain.Parking, 0)
	for rows.Next() {
		parking, err := scanParking(rows)
		if err != nil {
			return nil, err
		}
		parkings = append(parkings, parking)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parkings, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages parking registrations in every dormitory), along
// with their role ID so callers can also check role-level dormitory grants.
func (r *Repository) dormitoryScope(ctx context.Context, userID uuid.UUID) (full bool, roleID uuid.UUID, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT r.full_dormitory_access, r.id
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, userID).Scan(&full, &roleID)
	if err != nil {
		return false, uuid.Nil, err
	}
	return full, roleID, nil
}

// ensureRoomAccess confirms the room exists and the requester may assign
// parking registrations to it, based on access to its parent dormitory.
func (r *Repository) ensureRoomAccess(ctx context.Context, roomID, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM rooms rm
		WHERE rm.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rm.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rm.dormitory_id AND rd.role_id = $4
		))
	`, roomID, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return parkingdomain.ErrRoomNotFound
		}
		return err
	}
	return nil
}

// ensureParkingAccess confirms the parking registration exists and the
// requester may act on it. Registrations without a room are only visible to
// roles with full dormitory access. Both a missing registration and a
// missing grant surface as ErrParkingNotFound so scoped-out callers can't
// distinguish the two.
func (r *Repository) ensureParkingAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM parking_registrations p
		LEFT JOIN rooms rm ON rm.id = p.room_id
		WHERE p.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rm.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rm.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return parkingdomain.ErrParkingNotFound
		}
		return err
	}
	return nil
}
