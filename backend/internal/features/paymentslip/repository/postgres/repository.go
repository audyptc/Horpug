package postgres

import (
	"context"
	"errors"
	"time"

	slipdomain "apihorpug/internal/features/paymentslip/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const selectSlip = `
	SELECT s.id, s.invoice_id, COALESCE(i.invoice_no, ''), s.tenant_id, t.first_name || ' ' || t.last_name, rm.room_number, d.id, d.name,
		i.period_year, i.period_month, i.total_amount::float8,
		GREATEST(i.total_amount - COALESCE((SELECT SUM(p.total_amount) FROM payments p WHERE p.invoice_id = i.id AND p.status = 'active'), 0), 0)::float8,
		s.amount::float8, s.transfer_date, s.note, s.file_mime, s.file_size, s.status, s.reject_reason,
		s.payment_id, COALESCE(pay.receipt_no, ''), s.reviewed_at, s.created_at, s.file_key, t.line_user_id
	FROM payment_slips s
	JOIN invoices i ON i.id = s.invoice_id
	JOIN contracts c ON c.id = i.contract_id
	JOIN rooms rm ON rm.id = c.room_id
	JOIN dormitories d ON d.id = rm.dormitory_id
	JOIN tenants t ON t.id = s.tenant_id
	LEFT JOIN payments pay ON pay.id = s.payment_id`

func scanSlip(row pgx.CollectableRow) (slipdomain.Slip, error) {
	var s slipdomain.Slip
	err := row.Scan(&s.ID, &s.InvoiceID, &s.InvoiceNo, &s.TenantID, &s.TenantName, &s.RoomNumber, &s.DormitoryID, &s.DormitoryName,
		&s.PeriodYear, &s.PeriodMonth, &s.InvoiceTotal, &s.Outstanding,
		&s.Amount, &s.TransferDate, &s.Note, &s.FileMime, &s.FileSize, &s.Status, &s.RejectReason,
		&s.PaymentID, &s.ReceiptNo, &s.ReviewedAt, &s.CreatedAt, &s.FileKey, &s.TenantLineUserID)
	return s, err
}

func (r *Repository) query(ctx context.Context, where string, args ...any) ([]slipdomain.Slip, error) {
	rows, err := r.db.Query(ctx, selectSlip+" WHERE "+where, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanSlip)
}

func (r *Repository) one(ctx context.Context, where string, args ...any) (slipdomain.Slip, error) {
	slips, err := r.query(ctx, where, args...)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	if len(slips) == 0 {
		return slipdomain.Slip{}, slipdomain.ErrSlipNotFound
	}
	return slips[0], nil
}

// ---- tenant side: every query is keyed by the tenant's own id.

// PayableInvoiceForTenant returns what is still owed on an unpaid or overdue
// invoice of the tenant's; anything else is ErrInvoiceNotPayable.
func (r *Repository) PayableInvoiceForTenant(ctx context.Context, tenantID, invoiceID uuid.UUID) (float64, error) {
	var outstanding float64
	err := r.db.QueryRow(ctx, `
		SELECT GREATEST(i.total_amount - COALESCE((SELECT SUM(p.total_amount) FROM payments p WHERE p.invoice_id = i.id AND p.status = 'active'), 0), 0)::float8
		FROM invoices i JOIN contracts c ON c.id = i.contract_id
		WHERE i.id = $1 AND c.tenant_id = $2 AND i.status IN ('unpaid', 'overdue')
	`, invoiceID, tenantID).Scan(&outstanding)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, slipdomain.ErrInvoiceNotPayable
	}
	return outstanding, err
}

func (r *Repository) Create(ctx context.Context, s slipdomain.Slip) (slipdomain.Slip, error) {
	id := uuid.New()
	if _, err := r.db.Exec(ctx, `
		INSERT INTO payment_slips (id, invoice_id, tenant_id, amount, transfer_date, note, file_key, file_mime, file_size)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, s.InvoiceID, s.TenantID, s.Amount, s.TransferDate, s.Note, s.FileKey, s.FileMime, s.FileSize); err != nil {
		return slipdomain.Slip{}, err
	}
	return r.one(ctx, "s.id = $1", id)
}

func (r *Repository) ListForTenantInvoice(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]slipdomain.Slip, error) {
	return r.query(ctx, "s.tenant_id = $1 AND s.invoice_id = $2 ORDER BY s.created_at DESC", tenantID, invoiceID)
}

func (r *Repository) CancelByTenant(ctx context.Context, tenantID, id uuid.UUID) (slipdomain.Slip, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE payment_slips SET status = 'cancelled' WHERE id = $1 AND tenant_id = $2 AND status = 'pending'
	`, id, tenantID)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	slip, err := r.one(ctx, "s.id = $1 AND s.tenant_id = $2", id, tenantID)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	if tag.RowsAffected() == 0 {
		return slipdomain.Slip{}, slipdomain.ErrNotPending
	}
	return slip, nil
}

// ---- staff side: scoped to the dormitories the requester manages.

// staffScope limits slips to dormitories the requester manages; it binds
// $1 full access, $2 requester id and $3 role id (see scopeArgs).
const staffScope = `($1 OR d.id IN (
	SELECT dormitory_id FROM user_dormitories WHERE user_id = $2
	UNION
	SELECT dormitory_id FROM role_dormitories WHERE role_id = $3
))`

func (r *Repository) scopeArgs(ctx context.Context, requesterID uuid.UUID) ([]any, error) {
	var full bool
	var roleID uuid.UUID
	if err := r.db.QueryRow(ctx, `
		SELECT r.full_dormitory_access, r.id FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = $1
	`, requesterID).Scan(&full, &roleID); err != nil {
		return nil, err
	}
	return []any{full, requesterID, roleID}, nil
}

// ListForStaff lists slips with the given status, oldest first so the queue
// is worked in order, up to limit.
func (r *Repository) ListForStaff(ctx context.Context, requesterID uuid.UUID, status slipdomain.Status, limit int) ([]slipdomain.Slip, error) {
	args, err := r.scopeArgs(ctx, requesterID)
	if err != nil {
		return nil, err
	}
	order := "ASC"
	if status != slipdomain.StatusPending {
		order = "DESC"
	}
	return r.query(ctx, staffScope+" AND s.status = $4 ORDER BY s.created_at "+order+" LIMIT $5", append(args, status, limit)...)
}

func (r *Repository) GetForStaff(ctx context.Context, id, requesterID uuid.UUID) (slipdomain.Slip, error) {
	args, err := r.scopeArgs(ctx, requesterID)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	return r.one(ctx, staffScope+" AND s.id = $4", append(args, id)...)
}

// Claim marks a pending slip approved by reviewerID before the payment is
// recorded, so a second reviewer (or a double click) can't approve it too.
// It reports false when the slip was no longer pending.
func (r *Repository) Claim(ctx context.Context, id, reviewerID uuid.UUID) (bool, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE payment_slips SET status = 'approved', reviewed_by = $2, reviewed_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, id, reviewerID)
	return tag.RowsAffected() == 1, err
}

// Unclaim puts a claimed slip back to pending when recording its payment
// failed.
func (r *Repository) Unclaim(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE payment_slips SET status = 'pending', reviewed_by = NULL, reviewed_at = NULL
		WHERE id = $1 AND status = 'approved' AND payment_id IS NULL
	`, id)
	return err
}

func (r *Repository) SetPayment(ctx context.Context, id, paymentID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE payment_slips SET payment_id = $2 WHERE id = $1`, id, paymentID)
	return err
}

func (r *Repository) Reject(ctx context.Context, id, reviewerID uuid.UUID, reason string) (bool, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE payment_slips SET status = 'rejected', reject_reason = $3, reviewed_by = $2, reviewed_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, id, reviewerID, reason)
	return tag.RowsAffected() == 1, err
}

// today is the Bangkok calendar date (the session time zone).
func (r *Repository) Today(ctx context.Context) (time.Time, error) {
	var d time.Time
	err := r.db.QueryRow(ctx, `SELECT CURRENT_DATE`).Scan(&d)
	return d, err
}
