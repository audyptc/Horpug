package postgres

import (
	"context"

	invoicedomain "apihorpug/internal/features/invoice/domain"

	"github.com/google/uuid"
)

// ListDueReminders returns overdue invoices whose tenant has a linked LINE
// account, in dormitories with reminders switched on, that haven't been
// reminded in the last 7 days. Days are compared as Bangkok calendar dates
// (the session time zone, see database.NewPostgres), so the reminder lands on
// the same weekday each week instead of drifting with the hourly sweep.
func (r *Repository) ListDueReminders(ctx context.Context) ([]invoicedomain.ReminderCandidate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT i.id, COALESCE(i.invoice_no, ''), d.id, d.name, d.promptpay_id, rm.room_number, t.line_user_id,
			i.period_year, i.period_month, i.due_date, i.total_amount::float8,
			COALESCE((SELECT SUM(p.total_amount) FROM payments p WHERE p.invoice_id = i.id AND p.status = 'active'), 0)::float8,
			i.reminder_count
		FROM invoices i
		JOIN contracts c ON c.id = i.contract_id
		JOIN tenants t ON t.id = c.tenant_id
		JOIN rooms rm ON rm.id = c.room_id
		JOIN dormitories d ON d.id = rm.dormitory_id
		WHERE i.status = 'overdue'
		AND d.overdue_reminder_enabled
		AND t.line_user_id <> ''
		AND (i.last_reminder_at IS NULL OR i.last_reminder_at::date <= CURRENT_DATE - 7)
		ORDER BY i.due_date ASC, i.id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]invoicedomain.ReminderCandidate, 0)
	for rows.Next() {
		var c invoicedomain.ReminderCandidate
		if err := rows.Scan(&c.InvoiceID, &c.InvoiceNo, &c.DormitoryID, &c.DormitoryName, &c.PromptPayID, &c.RoomNumber, &c.TenantLineUserID,
			&c.PeriodYear, &c.PeriodMonth, &c.DueDate, &c.TotalAmount, &c.PaidAmount, &c.ReminderCount); err != nil {
			return nil, err
		}
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}

// RecordReminder logs a reminder attempt and starts the invoice's next 7-day
// wait. Only delivered reminders count towards reminder_count.
func (r *Repository) RecordReminder(ctx context.Context, invoiceID uuid.UUID, outcome invoicedomain.ReminderOutcome) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO invoice_reminders (id, invoice_id, outcome) VALUES ($1, $2, $3)
	`, uuid.New(), invoiceID, outcome); err != nil {
		return err
	}

	sent := 0
	if outcome == invoicedomain.ReminderSent {
		sent = 1
	}
	if _, err := tx.Exec(ctx, `
		UPDATE invoices SET last_reminder_at = NOW(), reminder_count = reminder_count + $2 WHERE id = $1
	`, invoiceID, sent); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
