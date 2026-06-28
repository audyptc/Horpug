package repository

import (
	"context"
	"fmt"
	"time"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type WaterMeterRepo struct {
	db *database.DB
}

func NewWaterMeterRepo(db *database.DB) *WaterMeterRepo {
	return &WaterMeterRepo{db: db}
}

const waterMeterDetailSelect = `
	SELECT
		w.id, w.room_id, w.billing_type, w.billing_month, w.reading_date,
		w.previous_reading, w.current_reading, w.unit_price, w.flat_amount,
		w.note, w.created_at, w.updated_at,
		COALESCE(w.created_by::text, ''), COALESCE(w.updated_by::text, ''), COALESCE(ub.full_name, ''),
		r.room_number,
		CASE WHEN w.billing_type = 'meter' THEN (w.current_reading - w.previous_reading) ELSE NULL END AS unit_used,
		CASE WHEN w.billing_type = 'flat' THEN w.flat_amount
		     ELSE (w.current_reading - w.previous_reading) * w.unit_price
		END AS total_amount
	FROM water_meter_readings w
	JOIN rooms r ON r.id = w.room_id
	LEFT JOIN users ub ON ub.id = w.updated_by`

func scanWaterMeterDetail(row pgx.Row) (*domain.WaterMeterDetail, error) {
	d := &domain.WaterMeterDetail{}
	err := row.Scan(
		&d.ID, &d.RoomID, &d.BillingType, &d.BillingMonth, &d.ReadingDate,
		&d.PreviousReading, &d.CurrentReading, &d.UnitPrice, &d.FlatAmount,
		&d.Note, &d.CreatedAt, &d.UpdatedAt,
		&d.CreatedBy, &d.UpdatedBy, &d.UpdatedByName,
		&d.RoomNumber, &d.UnitUsed, &d.TotalAmount,
	)
	return d, err
}

func (r *WaterMeterRepo) FindByID(ctx context.Context, id string) (*domain.WaterMeter, error) {
	m := &domain.WaterMeter{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, room_id, billing_type, billing_month, reading_date,
		       previous_reading, current_reading, unit_price, flat_amount, note, created_at, updated_at,
		       COALESCE(created_by::text, ''), COALESCE(updated_by::text, '')
		FROM water_meter_readings WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&m.ID, &m.RoomID, &m.BillingType, &m.BillingMonth, &m.ReadingDate,
			&m.PreviousReading, &m.CurrentReading, &m.UnitPrice, &m.FlatAmount,
			&m.Note, &m.CreatedAt, &m.UpdatedAt,
			&m.CreatedBy, &m.UpdatedBy)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("water meter reading not found: %w", domain.ErrNotFound)
	}
	return m, err
}

func (r *WaterMeterRepo) FindDetailByID(ctx context.Context, id string) (*domain.WaterMeterDetail, error) {
	row := r.db.Pool.QueryRow(ctx, waterMeterDetailSelect+` WHERE w.id = $1 AND w.deleted_at IS NULL`, id)
	d, err := scanWaterMeterDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("water meter reading not found: %w", domain.ErrNotFound)
	}
	return d, err
}

func (r *WaterMeterRepo) FindLatestByRoomID(ctx context.Context, roomID string, billingMonth *time.Time) (*domain.WaterMeterDetail, error) {
	q := waterMeterDetailSelect + ` WHERE w.room_id = $1 AND w.deleted_at IS NULL`
	args := []any{roomID}
	if billingMonth != nil {
		q += ` AND DATE_TRUNC('month', w.billing_month) = DATE_TRUNC('month', $2::date)`
		args = append(args, *billingMonth)
	}
	q += ` ORDER BY w.reading_date DESC, w.created_at DESC LIMIT 1`
	row := r.db.Pool.QueryRow(ctx, q, args...)
	d, err := scanWaterMeterDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("water meter reading not found: %w", domain.ErrNotFound)
	}
	return d, err
}

func (r *WaterMeterRepo) List(ctx context.Context, limit, offset int) ([]*domain.WaterMeterDetail, error) {
	rows, err := r.db.Pool.Query(ctx,
		waterMeterDetailSelect+` WHERE w.deleted_at IS NULL ORDER BY w.reading_date DESC, w.created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.WaterMeterDetail
	for rows.Next() {
		d, err := scanWaterMeterDetail(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	if list == nil {
		list = []*domain.WaterMeterDetail{}
	}
	return list, rows.Err()
}

func (r *WaterMeterRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM water_meter_readings WHERE deleted_at IS NULL`).Scan(&total)
	return total, err
}

func (r *WaterMeterRepo) Create(ctx context.Context, m *domain.WaterMeter) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO water_meter_readings
		    (id, room_id, billing_type, billing_month, reading_date, previous_reading, current_reading, unit_price, flat_amount, note, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NULLIF($11, '')::uuid, NULLIF($11, '')::uuid)`,
		m.ID, m.RoomID, m.BillingType, m.BillingMonth, m.ReadingDate,
		m.PreviousReading, m.CurrentReading, m.UnitPrice, m.FlatAmount, m.Note, m.CreatedBy)
	return err
}

func (r *WaterMeterRepo) Update(ctx context.Context, m *domain.WaterMeter) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE water_meter_readings
		SET billing_type = $2, billing_month = $3, reading_date = $4, previous_reading = $5, current_reading = $6,
		    unit_price = $7, flat_amount = $8, note = $9, updated_by = NULLIF($10, '')::uuid, updated_at = NOW()
		WHERE id = $1`,
		m.ID, m.BillingType, m.BillingMonth, m.ReadingDate,
		m.PreviousReading, m.CurrentReading, m.UnitPrice, m.FlatAmount, m.Note, m.UpdatedBy)
	return err
}

func (r *WaterMeterRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE water_meter_readings SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
