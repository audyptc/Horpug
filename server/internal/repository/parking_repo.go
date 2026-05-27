package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type ParkingRepo struct {
	db *database.DB
}

func NewParkingRepo(db *database.DB) *ParkingRepo {
	return &ParkingRepo{db: db}
}

func scanParkingSlot(row pgx.Row) (*domain.ParkingSlot, error) {
	p := &domain.ParkingSlot{}
	err := row.Scan(
		&p.ID, &p.SlotNumber, &p.VehicleType, &p.Status,
		&p.TenantID, &p.LicensePlate, &p.MonthlyFee, &p.StartDate,
		&p.Note, &p.CreatedAt, &p.UpdatedAt,
		&p.TenantFirstName, &p.TenantLastName,
	)
	return p, err
}

const selectParkingSlot = `
	SELECT
		ps.id, ps.slot_number, ps.vehicle_type, ps.status,
		ps.tenant_id, ps.license_plate, ps.monthly_fee, ps.start_date,
		ps.note, ps.created_at, ps.updated_at,
		COALESCE(t.first_name, '') AS tenant_first_name,
		COALESCE(t.last_name, '')  AS tenant_last_name
	FROM parking_slots ps
	LEFT JOIN tenants t ON t.id = ps.tenant_id`

func (r *ParkingRepo) FindByID(ctx context.Context, id string) (*domain.ParkingSlot, error) {
	row := r.db.Pool.QueryRow(ctx, selectParkingSlot+` WHERE ps.id = $1`, id)
	p, err := scanParkingSlot(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("parking slot not found: %w", domain.ErrNotFound)
	}
	return p, err
}

func (r *ParkingRepo) List(ctx context.Context, limit, offset int) ([]*domain.ParkingSlot, error) {
	rows, err := r.db.Pool.Query(ctx, selectParkingSlot+`
		ORDER BY ps.slot_number ASC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.ParkingSlot
	for rows.Next() {
		p, err := scanParkingSlot(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	if list == nil {
		list = []*domain.ParkingSlot{}
	}
	return list, rows.Err()
}

func (r *ParkingRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM parking_slots`).Scan(&total)
	return total, err
}

func (r *ParkingRepo) Create(ctx context.Context, p *domain.ParkingSlot) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO parking_slots (id, slot_number, vehicle_type, status, tenant_id, license_plate, monthly_fee, start_date, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.SlotNumber, p.VehicleType, p.Status,
		p.TenantID, p.LicensePlate, p.MonthlyFee, p.StartDate, p.Note)
	return err
}

func (r *ParkingRepo) Update(ctx context.Context, p *domain.ParkingSlot) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE parking_slots
		SET slot_number=$2, vehicle_type=$3, status=$4, tenant_id=$5,
		    license_plate=$6, monthly_fee=$7, start_date=$8, note=$9, updated_at=NOW()
		WHERE id=$1`,
		p.ID, p.SlotNumber, p.VehicleType, p.Status,
		p.TenantID, p.LicensePlate, p.MonthlyFee, p.StartDate, p.Note)
	return err
}

func (r *ParkingRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM parking_slots WHERE id = $1`, id)
	return err
}
