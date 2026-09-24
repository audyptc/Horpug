package postgres

import (
	"context"
	"errors"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ListGenerationCandidates returns the dormitory's active contracts that had
// started by the end of the period, ordered by room, flagged with whether the
// period is already billed and whether its meter readings are in. A dormitory
// the requester doesn't manage surfaces as ErrDormitoryNotFound.
func (r *Repository) ListGenerationCandidates(ctx context.Context, requesterID, dormitoryID uuid.UUID, periodYear, periodMonth int) ([]invoicedomain.GenerationCandidate, error) {
	if err := r.ensureDormitoryAccess(ctx, dormitoryID, requesterID); err != nil {
		return nil, err
	}

	periodStart := time.Date(periodYear, time.Month(periodMonth), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	rows, err := r.db.Query(ctx, `
		SELECT c.id, t.first_name || ' ' || t.last_name, rm.id, rm.room_number, c.rent_price, c.end_date,
			EXISTS (
				SELECT 1 FROM invoices i
				WHERE i.contract_id = c.id AND i.period_year = $2 AND i.period_month = $3
			),
			EXISTS (
				SELECT 1 FROM electricity_meters m
				WHERE m.room_id = rm.id AND m.reading_date >= $4 AND m.reading_date < $5
			),
			EXISTS (
				SELECT 1 FROM water_meters m
				WHERE m.room_id = rm.id AND m.reading_date >= $4 AND m.reading_date < $5
			),
			(c.end_date IS NOT NULL AND c.end_date < $4)
		FROM contracts c
		JOIN rooms rm ON rm.id = c.room_id
		JOIN tenants t ON t.id = c.tenant_id
		WHERE rm.dormitory_id = $1 AND c.status = 'active' AND c.start_date < $5
		ORDER BY rm.room_number ASC
	`, dormitoryID, periodYear, periodMonth, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]invoicedomain.GenerationCandidate, 0)
	for rows.Next() {
		var c invoicedomain.GenerationCandidate
		if err := rows.Scan(&c.ContractID, &c.TenantName, &c.RoomID, &c.RoomNumber, &c.RentPrice, &c.EndDate,
			&c.AlreadyInvoiced, &c.HasElectricity, &c.HasWater, &c.EndedBeforePeriod); err != nil {
			return nil, err
		}
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}

// MarkOverdue moves unpaid invoices past their due date to overdue, and moves
// overdue ones whose due date was pushed back to today or later back to
// unpaid. "Today" is the Bangkok calendar date, since due dates are local.
// Paid and cancelled invoices are never touched.
func (r *Repository) MarkOverdue(ctx context.Context) (markedOverdue, revertedUnpaid int64, err error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE invoices SET status = 'overdue', updated_at = NOW()
		WHERE status = 'unpaid' AND due_date < (NOW() AT TIME ZONE 'Asia/Bangkok')::date
	`)
	if err != nil {
		return 0, 0, err
	}
	markedOverdue = tag.RowsAffected()

	tag, err = r.db.Exec(ctx, `
		UPDATE invoices SET status = 'unpaid', updated_at = NOW()
		WHERE status = 'overdue' AND due_date >= (NOW() AT TIME ZONE 'Asia/Bangkok')::date
	`)
	if err != nil {
		return markedOverdue, 0, err
	}
	return markedOverdue, tag.RowsAffected(), nil
}

func (r *Repository) ensureDormitoryAccess(ctx context.Context, dormitoryID, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM dormitories d
		WHERE d.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = d.id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = d.id AND rd.role_id = $4
		))
	`, dormitoryID, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicedomain.ErrDormitoryNotFound
		}
		return err
	}
	return nil
}
