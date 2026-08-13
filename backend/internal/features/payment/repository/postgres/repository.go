package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	paymentdomain "apihorpug/internal/features/payment/domain"
	paymentusecase "apihorpug/internal/features/payment/usecase"

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

const selectPaymentColumns = `
	p.id, p.invoice_id, c.tenant_id, t.first_name, t.last_name,
	c.room_id, rm.room_number, rm.dormitory_id, d.name,
	p.amount, p.payment_method, p.payment_date, p.reference_no, p.note,
	p.created_by, p.created_at
`

const paymentFromJoins = `
	FROM payments p
	JOIN invoices i ON i.id = p.invoice_id
	JOIN contracts c ON c.id = i.contract_id
	JOIN tenants t ON t.id = c.tenant_id
	JOIN rooms rm ON rm.id = c.room_id
	JOIN dormitories d ON d.id = rm.dormitory_id
`

func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters paymentusecase.ListFilters, argIdx *int, args *[]any) []string {
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
	if filters.InvoiceID != nil {
		conditions = append(conditions, fmt.Sprintf(`p.invoice_id = $%d`, *argIdx))
		*args = append(*args, *filters.InvoiceID)
		*argIdx++
	}
	if filters.ContractID != nil {
		conditions = append(conditions, fmt.Sprintf(`i.contract_id = $%d`, *argIdx))
		*args = append(*args, *filters.ContractID)
		*argIdx++
	}
	if filters.RoomID != nil {
		conditions = append(conditions, fmt.Sprintf(`c.room_id = $%d`, *argIdx))
		*args = append(*args, *filters.RoomID)
		*argIdx++
	}
	if filters.DormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id = $%d`, *argIdx))
		*args = append(*args, *filters.DormitoryID)
		*argIdx++
	}
	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf(`c.tenant_id = $%d`, *argIdx))
		*args = append(*args, *filters.TenantID)
		*argIdx++
	}
	if filters.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf(`p.payment_date >= $%d`, *argIdx))
		*args = append(*args, *filters.DateFrom)
		*argIdx++
	}
	if filters.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf(`p.payment_date <= $%d`, *argIdx))
		*args = append(*args, *filters.DateTo)
		*argIdx++
	}

	return conditions
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters paymentusecase.ListFilters) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT COUNT(*) ` + paymentFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters paymentusecase.ListFilters, limit, offset int) ([]paymentdomain.Payment, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT ` + selectPaymentColumns + paymentFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY p.payment_date DESC, p.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPayments(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (paymentdomain.Payment, error) {
	if err := r.ensurePaymentAccess(ctx, id, requesterID); err != nil {
		return paymentdomain.Payment{}, err
	}

	payment, err := r.loadPaymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return paymentdomain.Payment{}, paymentdomain.ErrPaymentNotFound
		}
		return paymentdomain.Payment{}, err
	}

	return payment, nil
}

// Create records a payment against an invoice and, once the invoice's
// payments sum to its total_amount or more, marks the invoice paid. It runs
// in a transaction so the payment row and invoice status stay consistent.
func (r *Repository) Create(ctx context.Context, input paymentusecase.CreateInput) (paymentdomain.Payment, error) {
	invoiceTotal, invoiceStatus, err := r.loadInvoiceForPayment(ctx, input.InvoiceID, input.CreatedBy)
	if err != nil {
		return paymentdomain.Payment{}, err
	}
	if invoiceStatus == "cancelled" {
		return paymentdomain.Payment{}, paymentdomain.ErrInvoiceCancelled
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return paymentdomain.Payment{}, err
	}
	defer tx.Rollback(ctx)

	id := uuid.New()
	if _, err := tx.Exec(ctx, `
		INSERT INTO payments (id, invoice_id, amount, payment_method, payment_date, reference_no, note, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, input.InvoiceID, input.Amount, input.PaymentMethod, input.PaymentDate, input.ReferenceNo, input.Note, input.CreatedBy); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return paymentdomain.Payment{}, paymentdomain.ErrInvoiceNotFound
		}
		return paymentdomain.Payment{}, err
	}

	var totalPaid float64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM payments WHERE invoice_id = $1`, input.InvoiceID).Scan(&totalPaid); err != nil {
		return paymentdomain.Payment{}, err
	}

	if totalPaid >= invoiceTotal && invoiceStatus != "paid" {
		if _, err := tx.Exec(ctx, `UPDATE invoices SET status = 'paid', paid_at = NOW(), updated_at = NOW() WHERE id = $1`, input.InvoiceID); err != nil {
			return paymentdomain.Payment{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return paymentdomain.Payment{}, err
	}

	return r.loadPaymentByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensurePaymentAccess(ctx, id, requesterID); err != nil {
		return err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var invoiceID uuid.UUID
	if err := tx.QueryRow(ctx, `DELETE FROM payments WHERE id = $1 RETURNING invoice_id`, id).Scan(&invoiceID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return paymentdomain.ErrPaymentNotFound
		}
		return err
	}

	var invoiceTotal float64
	var invoiceStatus string
	if err := tx.QueryRow(ctx, `SELECT total_amount, status FROM invoices WHERE id = $1`, invoiceID).Scan(&invoiceTotal, &invoiceStatus); err != nil {
		return err
	}

	var totalPaid float64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM payments WHERE invoice_id = $1`, invoiceID).Scan(&totalPaid); err != nil {
		return err
	}

	if totalPaid < invoiceTotal && invoiceStatus == "paid" {
		if _, err := tx.Exec(ctx, `UPDATE invoices SET status = 'unpaid', paid_at = NULL, updated_at = NOW() WHERE id = $1`, invoiceID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) loadPaymentByID(ctx context.Context, id uuid.UUID) (paymentdomain.Payment, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectPaymentColumns+paymentFromJoins+` WHERE p.id = $1`, id)
	return scanPayment(row)
}

func scanPayment(row pgx.Row) (paymentdomain.Payment, error) {
	var payment paymentdomain.Payment
	var firstName, lastName string
	if err := row.Scan(
		&payment.ID,
		&payment.InvoiceID,
		&payment.TenantID,
		&firstName,
		&lastName,
		&payment.RoomID,
		&payment.RoomNumber,
		&payment.DormitoryID,
		&payment.DormitoryName,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.PaymentDate,
		&payment.ReferenceNo,
		&payment.Note,
		&payment.CreatedBy,
		&payment.CreatedAt,
	); err != nil {
		return paymentdomain.Payment{}, err
	}
	payment.TenantName = strings.TrimSpace(firstName + " " + lastName)
	return payment, nil
}

func scanPayments(rows pgx.Rows) ([]paymentdomain.Payment, error) {
	payments := make([]paymentdomain.Payment, 0)
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages payments in every dormitory), along with their
// role ID so callers can also check role-level dormitory grants.
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

// loadInvoiceForPayment confirms the invoice exists and the requester may
// record a payment against it, based on access to the dormitory of its
// contract's room, then returns the invoice's total amount and status. A nil
// requesterID (should not normally happen, since Create always supplies one)
// skips the dormitory-access check.
func (r *Repository) loadInvoiceForPayment(ctx context.Context, invoiceID uuid.UUID, requesterID *uuid.UUID) (totalAmount float64, status string, err error) {
	if requesterID == nil {
		err = r.db.QueryRow(ctx, `SELECT total_amount, status FROM invoices WHERE id = $1`, invoiceID).Scan(&totalAmount, &status)
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", paymentdomain.ErrInvoiceNotFound
		}
		return totalAmount, status, err
	}

	full, roleID, err := r.dormitoryScope(ctx, *requesterID)
	if err != nil {
		return 0, "", err
	}

	err = r.db.QueryRow(ctx, `
		SELECT i.total_amount, i.status
		FROM invoices i
		JOIN contracts c ON c.id = i.contract_id
		JOIN rooms rm ON rm.id = c.room_id
		WHERE i.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rm.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rm.dormitory_id AND rd.role_id = $4
		))
	`, invoiceID, full, *requesterID, roleID).Scan(&totalAmount, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", paymentdomain.ErrInvoiceNotFound
		}
		return 0, "", err
	}
	return totalAmount, status, nil
}

// ensurePaymentAccess confirms the payment exists and the requester may act
// on it, based on access to the dormitory of its invoice's contract's room.
// Both a missing payment and a missing grant surface as ErrPaymentNotFound so
// scoped-out callers can't distinguish the two.
func (r *Repository) ensurePaymentAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		JOIN contracts c ON c.id = i.contract_id
		JOIN rooms rm ON rm.id = c.room_id
		WHERE p.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rm.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rm.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return paymentdomain.ErrPaymentNotFound
		}
		return err
	}
	return nil
}
