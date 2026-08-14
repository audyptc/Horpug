package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	parceldomain "apihorpug/internal/features/parcel/domain"
	parcelusecase "apihorpug/internal/features/parcel/usecase"

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

const selectParcelColumns = `
	p.id, p.tenant_id, t.first_name || ' ' || t.last_name, p.room_id, rm.room_number, rm.dormitory_id, d.name,
	p.courier, p.tracking_number, p.status, p.received_date, p.note, p.created_by, p.updated_by, p.created_at, p.updated_at
`

const parcelFromJoins = `
	FROM parcels p
	JOIN tenants t ON t.id = p.tenant_id
	LEFT JOIN rooms rm ON rm.id = p.room_id
	LEFT JOIN dormitories d ON d.id = rm.dormitory_id
`

// buildScope restricts results to dormitories the requester manages. Rows
// with no room assigned have no dormitory to check against, so they only
// ever match for roles with full dormitory access.
func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters parcelusecase.ListFilters, argIdx *int, args *[]any) []string {
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
	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf(`p.status = $%d`, *argIdx))
		*args = append(*args, *filters.Status)
		*argIdx++
	}

	return conditions
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters parcelusecase.ListFilters) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT COUNT(*) ` + parcelFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters parcelusecase.ListFilters, limit, offset int) ([]parceldomain.Parcel, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT ` + selectParcelColumns + parcelFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY p.received_date DESC, p.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanParcels(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (parceldomain.Parcel, error) {
	if err := r.ensureParcelAccess(ctx, id, requesterID); err != nil {
		return parceldomain.Parcel{}, err
	}

	parcel, err := r.loadParcelByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return parceldomain.Parcel{}, parceldomain.ErrParcelNotFound
		}
		return parceldomain.Parcel{}, err
	}
	return parcel, nil
}

func (r *Repository) Create(ctx context.Context, input parcelusecase.CreateInput) (parceldomain.Parcel, error) {
	if input.RoomID != nil && input.CreatedBy != nil {
		if err := r.ensureRoomAccess(ctx, *input.RoomID, *input.CreatedBy); err != nil {
			return parceldomain.Parcel{}, err
		}
	}

	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		INSERT INTO parcels (id, tenant_id, room_id, courier, tracking_number, status, received_date, note, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
	`, id, input.TenantID, input.RoomID, input.Courier, input.TrackingNumber, input.Status, input.ReceivedDate, input.Note, input.CreatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			if pgErr.ConstraintName == "parcels_room_fkey" {
				return parceldomain.Parcel{}, parceldomain.ErrRoomNotFound
			}
			return parceldomain.Parcel{}, parceldomain.ErrTenantNotFound
		}
		return parceldomain.Parcel{}, err
	}

	return r.loadParcelByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input parcelusecase.UpdateInput) (parceldomain.Parcel, error) {
	if err := r.ensureParcelAccess(ctx, id, requesterID); err != nil {
		return parceldomain.Parcel{}, err
	}
	if input.RoomID != nil {
		if err := r.ensureRoomAccess(ctx, *input.RoomID, requesterID); err != nil {
			return parceldomain.Parcel{}, err
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
	if input.Courier != nil {
		setClauses = append(setClauses, fmt.Sprintf("courier = $%d", argIdx))
		args = append(args, *input.Courier)
		argIdx++
	}
	if input.TrackingNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("tracking_number = $%d", argIdx))
		args = append(args, *input.TrackingNumber)
		argIdx++
	}
	if input.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *input.Status)
		argIdx++
	}
	if input.ReceivedDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("received_date = $%d", argIdx))
		args = append(args, *input.ReceivedDate)
		argIdx++
	}
	if input.Note != nil {
		setClauses = append(setClauses, fmt.Sprintf("note = $%d", argIdx))
		args = append(args, *input.Note)
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
		query := fmt.Sprintf("UPDATE parcels SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			return parceldomain.Parcel{}, err
		}
	}

	return r.loadParcelByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureParcelAccess(ctx, id, requesterID); err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, `DELETE FROM parcels WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return parceldomain.ErrParcelNotFound
	}

	return nil
}

func (r *Repository) loadParcelByID(ctx context.Context, id uuid.UUID) (parceldomain.Parcel, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectParcelColumns+parcelFromJoins+` WHERE p.id = $1`, id)
	return scanParcel(row)
}

func scanParcel(row pgx.Row) (parceldomain.Parcel, error) {
	var parcel parceldomain.Parcel
	if err := row.Scan(
		&parcel.ID,
		&parcel.TenantID,
		&parcel.TenantName,
		&parcel.RoomID,
		&parcel.RoomNumber,
		&parcel.DormitoryID,
		&parcel.DormitoryName,
		&parcel.Courier,
		&parcel.TrackingNumber,
		&parcel.Status,
		&parcel.ReceivedDate,
		&parcel.Note,
		&parcel.CreatedBy,
		&parcel.UpdatedBy,
		&parcel.CreatedAt,
		&parcel.UpdatedAt,
	); err != nil {
		return parceldomain.Parcel{}, err
	}
	return parcel, nil
}

func scanParcels(rows pgx.Rows) ([]parceldomain.Parcel, error) {
	parcels := make([]parceldomain.Parcel, 0)
	for rows.Next() {
		parcel, err := scanParcel(rows)
		if err != nil {
			return nil, err
		}
		parcels = append(parcels, parcel)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parcels, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages parcels in every dormitory), along with their
// role ID so callers can also check role-level dormitory grants.
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

// ensureRoomAccess confirms the room exists and the requester may attach
// parcel records to it, based on access to its parent dormitory.
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
			return parceldomain.ErrRoomNotFound
		}
		return err
	}
	return nil
}

// ensureParcelAccess confirms the parcel record exists and the requester may
// act on it. Records without a room are only visible to roles with full
// dormitory access. Both a missing record and a missing grant surface as
// ErrParcelNotFound so scoped-out callers can't distinguish the two.
func (r *Repository) ensureParcelAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM parcels p
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
			return parceldomain.ErrParcelNotFound
		}
		return err
	}
	return nil
}
