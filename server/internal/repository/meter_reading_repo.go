package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type MeterReadingRepo struct {
	db *database.DB
}

func NewMeterReadingRepo(db *database.DB) *MeterReadingRepo {
	return &MeterReadingRepo{db: db}
}

const meterReadingDetailSelect = `
	SELECT
		m.id, m.room_id, m.meter_type, m.reading_date,
		m.previous_reading, m.current_reading, m.unit_price, m.note,
		m.created_at, m.updated_at,
		r.room_number,
		(m.current_reading - m.previous_reading) AS unit_used,
		(m.current_reading - m.previous_reading) * m.unit_price AS total_amount
	FROM meter_readings m
	JOIN rooms r ON r.id = m.room_id`

func scanMeterReadingDetail(row pgx.Row) (*domain.MeterReadingDetail, error) {
	d := &domain.MeterReadingDetail{}
	err := row.Scan(
		&d.ID, &d.RoomID, &d.MeterType, &d.ReadingDate,
		&d.PreviousReading, &d.CurrentReading, &d.UnitPrice, &d.Note,
		&d.CreatedAt, &d.UpdatedAt,
		&d.RoomNumber, &d.UnitUsed, &d.TotalAmount,
	)
	return d, err
}

func (r *MeterReadingRepo) FindByID(ctx context.Context, id string) (*domain.MeterReading, error) {
	m := &domain.MeterReading{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, room_id, meter_type, reading_date,
		       previous_reading, current_reading, unit_price, note, created_at, updated_at
		FROM meter_readings WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&m.ID, &m.RoomID, &m.MeterType, &m.ReadingDate,
			&m.PreviousReading, &m.CurrentReading, &m.UnitPrice, &m.Note,
			&m.CreatedAt, &m.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("meter reading not found: %w", domain.ErrNotFound)
	}
	return m, err
}

func (r *MeterReadingRepo) FindDetailByID(ctx context.Context, id string) (*domain.MeterReadingDetail, error) {
	row := r.db.Pool.QueryRow(ctx, meterReadingDetailSelect+` WHERE m.id = $1 AND m.deleted_at IS NULL`, id)
	d, err := scanMeterReadingDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("meter reading not found: %w", domain.ErrNotFound)
	}
	return d, err
}

func (r *MeterReadingRepo) List(ctx context.Context, limit, offset int) ([]*domain.MeterReadingDetail, error) {
	rows, err := r.db.Pool.Query(ctx,
		meterReadingDetailSelect+` WHERE m.deleted_at IS NULL ORDER BY m.reading_date DESC, m.created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.MeterReadingDetail
	for rows.Next() {
		d, err := scanMeterReadingDetail(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	if list == nil {
		list = []*domain.MeterReadingDetail{}
	}
	return list, rows.Err()
}

func (r *MeterReadingRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM meter_readings WHERE deleted_at IS NULL`).Scan(&total)
	return total, err
}

func (r *MeterReadingRepo) Create(ctx context.Context, m *domain.MeterReading) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO meter_readings (id, room_id, meter_type, reading_date, previous_reading, current_reading, unit_price, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		m.ID, m.RoomID, m.MeterType, m.ReadingDate, m.PreviousReading, m.CurrentReading, m.UnitPrice, m.Note)
	return err
}

func (r *MeterReadingRepo) Update(ctx context.Context, m *domain.MeterReading) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE meter_readings
		SET reading_date = $2, previous_reading = $3, current_reading = $4,
		    unit_price = $5, note = $6, updated_at = NOW()
		WHERE id = $1`,
		m.ID, m.ReadingDate, m.PreviousReading, m.CurrentReading, m.UnitPrice, m.Note)
	return err
}

func (r *MeterReadingRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE meter_readings SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
