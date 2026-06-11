package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type ElectricMeterRepo struct {
	db *database.DB
}

func NewElectricMeterRepo(db *database.DB) *ElectricMeterRepo {
	return &ElectricMeterRepo{db: db}
}

const electricMeterDetailSelect = `
	SELECT
		e.id, e.room_id, e.billing_type, e.reading_date,
		e.previous_reading, e.current_reading, e.unit_price, e.flat_amount,
		e.note, e.created_at, e.updated_at,
		r.room_number,
		CASE WHEN e.billing_type = 'meter' THEN (e.current_reading - e.previous_reading) ELSE NULL END AS unit_used,
		CASE WHEN e.billing_type = 'flat' THEN e.flat_amount
		     ELSE (e.current_reading - e.previous_reading) * e.unit_price
		END AS total_amount
	FROM electric_meter_readings e
	JOIN rooms r ON r.id = e.room_id`

func scanElectricMeterDetail(row pgx.Row) (*domain.ElectricMeterDetail, error) {
	d := &domain.ElectricMeterDetail{}
	err := row.Scan(
		&d.ID, &d.RoomID, &d.BillingType, &d.ReadingDate,
		&d.PreviousReading, &d.CurrentReading, &d.UnitPrice, &d.FlatAmount,
		&d.Note, &d.CreatedAt, &d.UpdatedAt,
		&d.RoomNumber, &d.UnitUsed, &d.TotalAmount,
	)
	return d, err
}

func (r *ElectricMeterRepo) FindByID(ctx context.Context, id string) (*domain.ElectricMeter, error) {
	m := &domain.ElectricMeter{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, room_id, billing_type, reading_date,
		       previous_reading, current_reading, unit_price, flat_amount, note, created_at, updated_at
		FROM electric_meter_readings WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&m.ID, &m.RoomID, &m.BillingType, &m.ReadingDate,
			&m.PreviousReading, &m.CurrentReading, &m.UnitPrice, &m.FlatAmount,
			&m.Note, &m.CreatedAt, &m.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("electric meter reading not found: %w", domain.ErrNotFound)
	}
	return m, err
}

func (r *ElectricMeterRepo) FindDetailByID(ctx context.Context, id string) (*domain.ElectricMeterDetail, error) {
	row := r.db.Pool.QueryRow(ctx, electricMeterDetailSelect+` WHERE e.id = $1 AND e.deleted_at IS NULL`, id)
	d, err := scanElectricMeterDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("electric meter reading not found: %w", domain.ErrNotFound)
	}
	return d, err
}

func (r *ElectricMeterRepo) List(ctx context.Context, limit, offset int) ([]*domain.ElectricMeterDetail, error) {
	rows, err := r.db.Pool.Query(ctx,
		electricMeterDetailSelect+` WHERE e.deleted_at IS NULL ORDER BY e.reading_date DESC, e.created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.ElectricMeterDetail
	for rows.Next() {
		d, err := scanElectricMeterDetail(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	if list == nil {
		list = []*domain.ElectricMeterDetail{}
	}
	return list, rows.Err()
}

func (r *ElectricMeterRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM electric_meter_readings WHERE deleted_at IS NULL`).Scan(&total)
	return total, err
}

func (r *ElectricMeterRepo) Create(ctx context.Context, m *domain.ElectricMeter) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO electric_meter_readings
		    (id, room_id, billing_type, reading_date, previous_reading, current_reading, unit_price, flat_amount, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		m.ID, m.RoomID, m.BillingType, m.ReadingDate,
		m.PreviousReading, m.CurrentReading, m.UnitPrice, m.FlatAmount, m.Note)
	return err
}

func (r *ElectricMeterRepo) Update(ctx context.Context, m *domain.ElectricMeter) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE electric_meter_readings
		SET billing_type = $2, reading_date = $3, previous_reading = $4, current_reading = $5,
		    unit_price = $6, flat_amount = $7, note = $8, updated_at = NOW()
		WHERE id = $1`,
		m.ID, m.BillingType, m.ReadingDate,
		m.PreviousReading, m.CurrentReading, m.UnitPrice, m.FlatAmount, m.Note)
	return err
}

func (r *ElectricMeterRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE electric_meter_readings SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
