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
		e.id, e.room_id, e.reading_date,
		e.previous_reading, e.current_reading, e.unit_price, e.note,
		e.created_at, e.updated_at,
		r.room_number,
		(e.current_reading - e.previous_reading) AS unit_used,
		(e.current_reading - e.previous_reading) * e.unit_price AS total_amount
	FROM electric_meter_readings e
	JOIN rooms r ON r.id = e.room_id`

func scanElectricMeterDetail(row pgx.Row) (*domain.ElectricMeterDetail, error) {
	d := &domain.ElectricMeterDetail{}
	err := row.Scan(
		&d.ID, &d.RoomID, &d.ReadingDate,
		&d.PreviousReading, &d.CurrentReading, &d.UnitPrice, &d.Note,
		&d.CreatedAt, &d.UpdatedAt,
		&d.RoomNumber, &d.UnitUsed, &d.TotalAmount,
	)
	return d, err
}

func (r *ElectricMeterRepo) FindByID(ctx context.Context, id string) (*domain.ElectricMeter, error) {
	m := &domain.ElectricMeter{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, room_id, reading_date,
		       previous_reading, current_reading, unit_price, note, created_at, updated_at
		FROM electric_meter_readings WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&m.ID, &m.RoomID, &m.ReadingDate,
			&m.PreviousReading, &m.CurrentReading, &m.UnitPrice, &m.Note,
			&m.CreatedAt, &m.UpdatedAt)
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
		INSERT INTO electric_meter_readings (id, room_id, reading_date, previous_reading, current_reading, unit_price, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		m.ID, m.RoomID, m.ReadingDate, m.PreviousReading, m.CurrentReading, m.UnitPrice, m.Note)
	return err
}

func (r *ElectricMeterRepo) Update(ctx context.Context, m *domain.ElectricMeter) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE electric_meter_readings
		SET reading_date = $2, previous_reading = $3, current_reading = $4,
		    unit_price = $5, note = $6, updated_at = NOW()
		WHERE id = $1`,
		m.ID, m.ReadingDate, m.PreviousReading, m.CurrentReading, m.UnitPrice, m.Note)
	return err
}

func (r *ElectricMeterRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE electric_meter_readings SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
