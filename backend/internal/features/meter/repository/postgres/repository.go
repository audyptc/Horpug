package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	meterdomain "apihorpug/internal/features/meter/domain"
	meterusecase "apihorpug/internal/features/meter/usecase"

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
	em.id, em.room_id, rm.room_number, rm.dormitory_id, d.name, em.billing_method,
	em.reading_date, em.previous_unit, em.current_unit, em.unit_used, em.price_per_unit, em.flat_amount, em.total_amount, em.note,
	em.created_by, em.updated_by, em.created_at, em.updated_at
`

const meterFromJoins = `
	FROM electricity_meters em
	JOIN rooms rm ON rm.id = em.room_id
	JOIN dormitories d ON d.id = rm.dormitory_id
`

func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters meterusecase.ListFilters, argIdx *int, args *[]any) []string {
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
		conditions = append(conditions, fmt.Sprintf(`em.room_id = $%d`, *argIdx))
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

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters meterusecase.ListFilters) (int64, error) {
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

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters meterusecase.ListFilters, limit, offset int) ([]meterdomain.Meter, error) {
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
	query += fmt.Sprintf(` ORDER BY em.reading_date DESC, rm.room_number ASC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMeters(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (meterdomain.Meter, error) {
	if err := r.ensureMeterAccess(ctx, id, requesterID); err != nil {
		return meterdomain.Meter{}, err
	}

	meter, err := r.loadMeterByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return meterdomain.Meter{}, meterdomain.ErrMeterNotFound
		}
		return meterdomain.Meter{}, err
	}
	return meter, nil
}

func (r *Repository) Create(ctx context.Context, input meterusecase.CreateInput) (meterdomain.Meter, error) {
	if input.CreatedBy != nil {
		if err := r.ensureRoomAccess(ctx, input.RoomID, *input.CreatedBy); err != nil {
			return meterdomain.Meter{}, err
		}
	}

	id := uuid.New()
	err := r.db.QueryRow(ctx, `
		INSERT INTO electricity_meters (id, room_id, billing_method, reading_date, previous_unit, current_unit, price_per_unit, flat_amount, note, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, id, input.RoomID, input.BillingMethod, input.ReadingDate, input.PreviousUnit, input.CurrentUnit, input.PricePerUnit, input.FlatAmount, input.Note, input.CreatedBy, input.CreatedBy).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return meterdomain.Meter{}, meterdomain.ErrMeterReadingExists
			}
			if pgErr.Code == "23503" {
				return meterdomain.Meter{}, meterdomain.ErrRoomNotFound
			}
			if pgErr.Code == "23514" {
				if pgErr.ConstraintName == "chk_electricity_meters_billing_method" {
					return meterdomain.Meter{}, meterdomain.ErrInvalidBillingMethod
				}
				if strings.HasPrefix(pgErr.ConstraintName, "chk_electricity_meters_flat_amount") {
					return meterdomain.Meter{}, meterdomain.ErrRequiredFlatAmount
				}
				return meterdomain.Meter{}, meterdomain.ErrInvalidMeterUnits
			}
		}
		return meterdomain.Meter{}, err
	}

	return r.loadMeterByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input meterusecase.UpdateInput) (meterdomain.Meter, error) {
	if err := r.ensureMeterAccess(ctx, id, requesterID); err != nil {
		return meterdomain.Meter{}, err
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
	if input.BillingMethod != nil && *input.BillingMethod == meterdomain.BillingMethodMetered {
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
		query := fmt.Sprintf("UPDATE electricity_meters SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					return meterdomain.Meter{}, meterdomain.ErrMeterReadingExists
				}
				if pgErr.Code == "23514" {
					if pgErr.ConstraintName == "chk_electricity_meters_billing_method" {
						return meterdomain.Meter{}, meterdomain.ErrInvalidBillingMethod
					}
					if strings.HasPrefix(pgErr.ConstraintName, "chk_electricity_meters_flat_amount") {
						return meterdomain.Meter{}, meterdomain.ErrRequiredFlatAmount
					}
					return meterdomain.Meter{}, meterdomain.ErrInvalidMeterUnits
				}
			}
			return meterdomain.Meter{}, err
		}
	}

	return r.loadMeterByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureMeterAccess(ctx, id, requesterID); err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, `DELETE FROM electricity_meters WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return meterdomain.ErrMeterNotFound
	}

	return nil
}

func (r *Repository) loadMeterByID(ctx context.Context, id uuid.UUID) (meterdomain.Meter, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectMeterColumns+meterFromJoins+` WHERE em.id = $1`, id)
	return scanMeter(row)
}

func scanMeter(row pgx.Row) (meterdomain.Meter, error) {
	var meter meterdomain.Meter
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
		&meter.Note,
		&meter.CreatedBy,
		&meter.UpdatedBy,
		&meter.CreatedAt,
		&meter.UpdatedAt,
	); err != nil {
		return meterdomain.Meter{}, err
	}
	return meter, nil
}

func scanMeters(rows pgx.Rows) ([]meterdomain.Meter, error) {
	meters := make([]meterdomain.Meter, 0)
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
			return meterdomain.ErrRoomNotFound
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
		SELECT 1 FROM electricity_meters em
		JOIN rooms rm ON rm.id = em.room_id
		WHERE em.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rm.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rm.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return meterdomain.ErrMeterNotFound
		}
		return err
	}
	return nil
}
