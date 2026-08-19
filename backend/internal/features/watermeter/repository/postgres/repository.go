package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	metdomain "apihorpug/internal/features/watermeter/domain"
	metusecase "apihorpug/internal/features/watermeter/usecase"

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

const selectMeterColumns = `
	wm.id, wm.room_id, rm.room_number, rm.dormitory_id, d.name, wm.billing_method,
	wm.reading_date, wm.previous_unit, wm.current_unit, wm.unit_used, wm.price_per_unit, wm.flat_amount, wm.total_amount,
	EXISTS (SELECT 1 FROM invoice_items ii WHERE ii.reference_id = wm.id AND ii.item_type = 'water') AS is_billed,
	wm.note, wm.created_by, wm.updated_by, wm.created_at, wm.updated_at
`

const meterFromJoins = `
	FROM water_meters wm
	JOIN rooms rm ON rm.id = wm.room_id
	JOIN dormitories d ON d.id = rm.dormitory_id
`

func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters metusecase.ListFilters, argIdx *int, args *[]any) []string {
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
	if filters.RoomID != nil {
		conditions = append(conditions, fmt.Sprintf(`wm.room_id = $%d`, *argIdx))
		*args = append(*args, *filters.RoomID)
		*argIdx++
	}
	if filters.DormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id = $%d`, *argIdx))
		*args = append(*args, *filters.DormitoryID)
		*argIdx++
	}

	return conditions
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters metusecase.ListFilters) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT COUNT(*) ` + meterFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters metusecase.ListFilters, limit, offset int) ([]metdomain.Meter, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT ` + selectMeterColumns + meterFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY wm.reading_date DESC, rm.room_number ASC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMeters(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (metdomain.Meter, error) {
	if err := r.ensureMeterAccess(ctx, id, requesterID); err != nil {
		return metdomain.Meter{}, err
	}

	meter, err := r.loadMeterByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return metdomain.Meter{}, metdomain.ErrMeterNotFound
		}
		return metdomain.Meter{}, err
	}
	return meter, nil
}

func (r *Repository) Create(ctx context.Context, input metusecase.CreateInput) (metdomain.Meter, error) {
	if input.CreatedBy != nil {
		if err := r.ensureRoomAccess(ctx, input.RoomID, *input.CreatedBy); err != nil {
			return metdomain.Meter{}, err
		}
	}

	duplicate, err := r.monthDuplicateExists(ctx, input.RoomID, input.ReadingDate, nil)
	if err != nil {
		return metdomain.Meter{}, err
	}
	if duplicate {
		return metdomain.Meter{}, metdomain.ErrMeterMonthExists
	}

	id := uuid.New()
	err = r.db.QueryRow(ctx, `
		INSERT INTO water_meters (id, room_id, billing_method, reading_date, previous_unit, current_unit, price_per_unit, flat_amount, note, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, id, input.RoomID, input.BillingMethod, input.ReadingDate, input.PreviousUnit, input.CurrentUnit, input.PricePerUnit, input.FlatAmount, input.Note, input.CreatedBy, input.CreatedBy).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return metdomain.Meter{}, metdomain.ErrMeterReadingExists
			}
			if pgErr.Code == "23503" {
				return metdomain.Meter{}, metdomain.ErrRoomNotFound
			}
			if pgErr.Code == "23514" {
				if pgErr.ConstraintName == "chk_water_meters_billing_method" {
					return metdomain.Meter{}, metdomain.ErrInvalidBillingMethod
				}
				if strings.HasPrefix(pgErr.ConstraintName, "chk_water_meters_flat_amount") {
					return metdomain.Meter{}, metdomain.ErrRequiredFlatAmount
				}
				return metdomain.Meter{}, metdomain.ErrInvalidMeterUnits
			}
		}
		return metdomain.Meter{}, err
	}

	return r.loadMeterByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input metusecase.UpdateInput) (metdomain.Meter, error) {
	if err := r.ensureMeterAccess(ctx, id, requesterID); err != nil {
		return metdomain.Meter{}, err
	}

	if input.ReadingDate != nil {
		var roomID uuid.UUID
		if err := r.db.QueryRow(ctx, `SELECT room_id FROM water_meters WHERE id = $1`, id).Scan(&roomID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return metdomain.Meter{}, metdomain.ErrMeterNotFound
			}
			return metdomain.Meter{}, err
		}

		duplicate, err := r.monthDuplicateExists(ctx, roomID, *input.ReadingDate, &id)
		if err != nil {
			return metdomain.Meter{}, err
		}
		if duplicate {
			return metdomain.Meter{}, metdomain.ErrMeterMonthExists
		}
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.BillingMethod != nil {
		setClauses = append(setClauses, fmt.Sprintf("billing_method = $%d", argIdx))
		args = append(args, *input.BillingMethod)
		argIdx++
	}
	if input.ReadingDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("reading_date = $%d", argIdx))
		args = append(args, *input.ReadingDate)
		argIdx++
	}
	if input.PreviousUnit != nil {
		setClauses = append(setClauses, fmt.Sprintf("previous_unit = $%d", argIdx))
		args = append(args, *input.PreviousUnit)
		argIdx++
	}
	if input.CurrentUnit != nil {
		setClauses = append(setClauses, fmt.Sprintf("current_unit = $%d", argIdx))
		args = append(args, *input.CurrentUnit)
		argIdx++
	}
	if input.PricePerUnit != nil {
		setClauses = append(setClauses, fmt.Sprintf("price_per_unit = $%d", argIdx))
		args = append(args, *input.PricePerUnit)
		argIdx++
	}
	// Switching to metered drops any flat_amount so it can't linger and
	// silently satisfy the "flat requires flat_amount" check if the row is
	// ever switched back to flat later without re-supplying it.
	if input.BillingMethod != nil && *input.BillingMethod == metdomain.BillingMethodMetered {
		setClauses = append(setClauses, "flat_amount = NULL")
	} else if input.FlatAmount != nil {
		setClauses = append(setClauses, fmt.Sprintf("flat_amount = $%d", argIdx))
		args = append(args, *input.FlatAmount)
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
		query := fmt.Sprintf("UPDATE water_meters SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					return metdomain.Meter{}, metdomain.ErrMeterReadingExists
				}
				if pgErr.Code == "23514" {
					if pgErr.ConstraintName == "chk_water_meters_billing_method" {
						return metdomain.Meter{}, metdomain.ErrInvalidBillingMethod
					}
					if strings.HasPrefix(pgErr.ConstraintName, "chk_water_meters_flat_amount") {
						return metdomain.Meter{}, metdomain.ErrRequiredFlatAmount
					}
					return metdomain.Meter{}, metdomain.ErrInvalidMeterUnits
				}
			}
			return metdomain.Meter{}, err
		}
	}

	return r.loadMeterByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureMeterAccess(ctx, id, requesterID); err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, `DELETE FROM water_meters WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return metdomain.ErrMeterNotFound
	}

	return nil
}

func (r *Repository) loadMeterByID(ctx context.Context, id uuid.UUID) (metdomain.Meter, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectMeterColumns+meterFromJoins+` WHERE wm.id = $1`, id)
	return scanMeter(row)
}

func scanMeter(row pgx.Row) (metdomain.Meter, error) {
	var meter metdomain.Meter
	if err := row.Scan(
		&meter.ID,
		&meter.RoomID,
		&meter.RoomNumber,
		&meter.DormitoryID,
		&meter.DormitoryName,
		&meter.BillingMethod,
		&meter.ReadingDate,
		&meter.PreviousUnit,
		&meter.CurrentUnit,
		&meter.UnitUsed,
		&meter.PricePerUnit,
		&meter.FlatAmount,
		&meter.TotalAmount,
		&meter.IsBilled,
		&meter.Note,
		&meter.CreatedBy,
		&meter.UpdatedBy,
		&meter.CreatedAt,
		&meter.UpdatedAt,
	); err != nil {
		return metdomain.Meter{}, err
	}
	return meter, nil
}

func scanMeters(rows pgx.Rows) ([]metdomain.Meter, error) {
	meters := make([]metdomain.Meter, 0)
	for rows.Next() {
		meter, err := scanMeter(rows)
		if err != nil {
			return nil, err
		}
		meters = append(meters, meter)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return meters, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages meter readings in every dormitory), along with
// their role ID so callers can also check role-level dormitory grants.
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

// monthDuplicateExists reports whether the room already has another meter
// reading whose reading_date falls in the same calendar month as
// readingDate. excludeID, when non-nil, excludes that row from the check
// (used on update, so a reading isn't flagged as a duplicate of itself).
func (r *Repository) monthDuplicateExists(ctx context.Context, roomID uuid.UUID, readingDate time.Time, excludeID *uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM water_meters
			WHERE room_id = $1
			AND date_trunc('month', reading_date) = date_trunc('month', $2::date)
			AND ($3::uuid IS NULL OR id != $3)
		)
	`, roomID, readingDate, excludeID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ensureRoomAccess confirms the room exists and the requester may record
// meter readings against it, based on access to its parent dormitory.
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
			return metdomain.ErrRoomNotFound
		}
		return err
	}
	return nil
}

// ensureMeterAccess confirms the meter reading exists and the requester may
// act on it, based on access to the dormitory of its room. Both a missing
// reading and a missing grant surface as ErrMeterNotFound so scoped-out
// callers can't distinguish the two.
func (r *Repository) ensureMeterAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM water_meters wm
		JOIN rooms rm ON rm.id = wm.room_id
		WHERE wm.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rm.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rm.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return metdomain.ErrMeterNotFound
		}
		return err
	}
	return nil
}
