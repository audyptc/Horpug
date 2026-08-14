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
	p.id, p.tenant_id, t.first_name || ' ' || t.last_name, p.vehicle_type, p.license_plate, p.parking_spot,
	p.created_by, p.updated_by, p.created_at, p.updated_at
`

const parkingFromJoins = `
	FROM parking_registrations p
	JOIN tenants t ON t.id = p.tenant_id
`

func (r *Repository) buildScope(filters parkingusecase.ListFilters, argIdx *int, args *[]any) []string {
	conditions := make([]string, 0)

	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf(`p.tenant_id = $%d`, *argIdx))
		*args = append(*args, *filters.TenantID)
		*argIdx++
	}
	if filters.VehicleType != nil {
		conditions = append(conditions, fmt.Sprintf(`p.vehicle_type = $%d`, *argIdx))
		*args = append(*args, *filters.VehicleType)
		*argIdx++
	}

	return conditions
}

func (r *Repository) Count(ctx context.Context, filters parkingusecase.ListFilters) (int64, error) {
	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(filters, &argIdx, &args)

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

func (r *Repository) List(ctx context.Context, filters parkingusecase.ListFilters, limit, offset int) ([]parkingdomain.Parking, error) {
	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(filters, &argIdx, &args)

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

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (parkingdomain.Parking, error) {
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
	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		INSERT INTO parking_registrations (id, tenant_id, vehicle_type, license_plate, parking_spot, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
	`, id, input.TenantID, input.VehicleType, input.LicensePlate, input.ParkingSpot, input.CreatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return parkingdomain.Parking{}, parkingdomain.ErrTenantNotFound
		}
		return parkingdomain.Parking{}, err
	}

	return r.loadParkingByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input parkingusecase.UpdateInput) (parkingdomain.Parking, error) {
	if err := r.ensureParkingExists(ctx, id); err != nil {
		return parkingdomain.Parking{}, err
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

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

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM parking_registrations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return parkingdomain.ErrParkingNotFound
	}

	return nil
}

func (r *Repository) ensureParkingExists(ctx context.Context, id uuid.UUID) error {
	var exists int
	if err := r.db.QueryRow(ctx, `SELECT 1 FROM parking_registrations WHERE id = $1`, id).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return parkingdomain.ErrParkingNotFound
		}
		return err
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
