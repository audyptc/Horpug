package repository

import (
	"context"
	"fmt"
	"time"

	"apigofiberhorpug/internal/feature/electricmeter/domain"
	"apigofiberhorpug/internal/platform/database"
	coredomain "apigofiberhorpug/internal/shared/domain"

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
		e.id, e.room_id, e.billing_type, e.billing_month, e.reading_date,
		e.previous_reading, e.current_reading, e.unit_price, e.flat_amount,
		e.note, e.created_at, e.updated_at,
		COALESCE(e.created_by::text, ''), COALESCE(e.updated_by::text, ''), COALESCE(ub.full_name, ''),
		r.room_number,
		CASE WHEN e.billing_type = 'meter' THEN (e.current_reading - e.previous_reading) ELSE NULL END AS unit_used,
		CASE WHEN e.billing_type = 'flat' THEN e.flat_amount
		     ELSE (e.current_reading - e.previous_reading) * e.unit_price
		END AS total_amount
	FROM electric_meter_readings e
	JOIN rooms r ON r.id = e.room_id
	LEFT JOIN users ub ON ub.id = e.updated_by`

func scanElectricMeterDetail(row pgx.Row) (*domain.ElectricMeterDetail, error) {
	d := &domain.ElectricMeterDetail{}
	err := row.Scan(
		&d.ID, &d.RoomID, &d.BillingType, &d.BillingMonth, &d.ReadingDate,
		&d.PreviousReading, &d.CurrentReading, &d.UnitPrice, &d.FlatAmount,
		&d.Note, &d.CreatedAt, &d.UpdatedAt,
		&d.CreatedBy, &d.UpdatedBy, &d.UpdatedByName,
		&d.RoomNumber, &d.UnitUsed, &d.TotalAmount,
	)
	return d, err
}

func (r *ElectricMeterRepo) FindByID(ctx context.Context, dormitoryID, id string) (*domain.ElectricMeter, error) {
	m := &domain.ElectricMeter{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT e.id, e.room_id, e.billing_type, e.billing_month, e.reading_date,
		       e.previous_reading, e.current_reading, e.unit_price, e.flat_amount, e.note, e.created_at, e.updated_at,
		       COALESCE(e.created_by::text, ''), COALESCE(e.updated_by::text, '')
		FROM electric_meter_readings e
		JOIN rooms r ON r.id = e.room_id
		WHERE e.id = $1 AND r.dormitory_id = $2 AND e.deleted_at IS NULL`, id, dormitoryID).
		Scan(&m.ID, &m.RoomID, &m.BillingType, &m.BillingMonth, &m.ReadingDate,
			&m.PreviousReading, &m.CurrentReading, &m.UnitPrice, &m.FlatAmount,
			&m.Note, &m.CreatedAt, &m.UpdatedAt,
			&m.CreatedBy, &m.UpdatedBy)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("electric meter reading not found: %w", coredomain.ErrNotFound)
	}
	return m, err
}

func (r *ElectricMeterRepo) FindDetailByID(ctx context.Context, dormitoryID, id string) (*domain.ElectricMeterDetail, error) {
	row := r.db.Pool.QueryRow(ctx, electricMeterDetailSelect+` WHERE e.id = $1 AND r.dormitory_id = $2 AND e.deleted_at IS NULL`, id, dormitoryID)
	d, err := scanElectricMeterDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("electric meter reading not found: %w", coredomain.ErrNotFound)
	}
	return d, err
}

func (r *ElectricMeterRepo) FindLatestByRoomID(ctx context.Context, dormitoryID, roomID string, billingMonth *time.Time) (*domain.ElectricMeterDetail, error) {
	q := electricMeterDetailSelect + ` WHERE e.room_id = $1 AND r.dormitory_id = $2 AND e.deleted_at IS NULL`
	args := []any{roomID, dormitoryID}
	if billingMonth != nil {
		q += ` AND DATE_TRUNC('month', e.billing_month) = DATE_TRUNC('month', $3::date)`
		args = append(args, *billingMonth)
	}
	q += ` ORDER BY e.reading_date DESC, e.created_at DESC LIMIT 1`
	row := r.db.Pool.QueryRow(ctx, q, args...)
	d, err := scanElectricMeterDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("electric meter reading not found: %w", coredomain.ErrNotFound)
	}
	return d, err
}

func (r *ElectricMeterRepo) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.ElectricMeterDetail, error) {
	rows, err := r.db.Pool.Query(ctx,
		electricMeterDetailSelect+` WHERE r.dormitory_id = $1 AND e.deleted_at IS NULL ORDER BY e.reading_date DESC, e.created_at DESC LIMIT $2 OFFSET $3`,
		dormitoryID, limit, offset)
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

func (r *ElectricMeterRepo) Count(ctx context.Context, dormitoryID string) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM electric_meter_readings e
		JOIN rooms r ON r.id = e.room_id
		WHERE r.dormitory_id = $1 AND e.deleted_at IS NULL`, dormitoryID).Scan(&total)
	return total, err
}

func (r *ElectricMeterRepo) Create(ctx context.Context, m *domain.ElectricMeter) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO electric_meter_readings
		    (id, room_id, billing_type, billing_month, reading_date, previous_reading, current_reading, unit_price, flat_amount, note, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NULLIF($11, '')::uuid, NULLIF($11, '')::uuid)`,
		m.ID, m.RoomID, m.BillingType, m.BillingMonth, m.ReadingDate,
		m.PreviousReading, m.CurrentReading, m.UnitPrice, m.FlatAmount, m.Note, m.CreatedBy)
	return err
}

func (r *ElectricMeterRepo) Update(ctx context.Context, dormitoryID string, m *domain.ElectricMeter) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE electric_meter_readings
		SET billing_type = $3, billing_month = $4, reading_date = $5, previous_reading = $6, current_reading = $7,
		    unit_price = $8, flat_amount = $9, note = $10, updated_by = NULLIF($11, '')::uuid, updated_at = NOW()
		FROM rooms r
		WHERE electric_meter_readings.id = $1 AND r.id = electric_meter_readings.room_id AND r.dormitory_id = $2`,
		m.ID, dormitoryID, m.BillingType, m.BillingMonth, m.ReadingDate,
		m.PreviousReading, m.CurrentReading, m.UnitPrice, m.FlatAmount, m.Note, m.UpdatedBy)
	return err
}

func (r *ElectricMeterRepo) Delete(ctx context.Context, dormitoryID, id string) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE electric_meter_readings SET deleted_at = NOW(), updated_at = NOW()
		FROM rooms r
		WHERE electric_meter_readings.id = $1 AND r.id = electric_meter_readings.room_id
		  AND r.dormitory_id = $2 AND electric_meter_readings.deleted_at IS NULL`, id, dormitoryID)
	return err
}
