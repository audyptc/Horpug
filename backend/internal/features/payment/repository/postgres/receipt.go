package postgres

import (
	"context"
	"errors"

	paymentdomain "apihorpug/internal/features/payment/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetReceipt loads a payment with its dormitory and invoice balance for
// printing. Access follows GetByID.
func (r *Repository) GetReceipt(ctx context.Context, id, requesterID uuid.UUID) (paymentdomain.Receipt, error) {
	payment, err := r.GetByID(ctx, id, requesterID)
	if err != nil {
		return paymentdomain.Receipt{}, err
	}

	return r.buildReceipt(ctx, payment)
}

// GetReceiptForTenant loads a receipt for the tenant self-service pages: only
// an active (not voided) payment on one of the tenant's own contracts.
// Anything else surfaces as ErrPaymentNotFound.
func (r *Repository) GetReceiptForTenant(ctx context.Context, id, tenantID uuid.UUID) (paymentdomain.Receipt, error) {
	var owned bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM payments p
			JOIN invoices i ON i.id = p.invoice_id
			JOIN contracts c ON c.id = i.contract_id
			WHERE p.id = $1 AND c.tenant_id = $2 AND p.status = 'active'
		)
	`, id, tenantID).Scan(&owned); err != nil {
		return paymentdomain.Receipt{}, err
	}
	if !owned {
		return paymentdomain.Receipt{}, paymentdomain.ErrPaymentNotFound
	}

	payment, err := r.loadPaymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return paymentdomain.Receipt{}, paymentdomain.ErrPaymentNotFound
		}
		return paymentdomain.Receipt{}, err
	}
	return r.buildReceipt(ctx, payment)
}

// buildReceipt adds the issuing dormitory and the invoice balance to a
// payment the caller has already been allowed to see.
func (r *Repository) buildReceipt(ctx context.Context, payment paymentdomain.Payment) (paymentdomain.Receipt, error) {
	id := payment.ID
	receipt := paymentdomain.Receipt{Payment: payment}
	err := r.db.QueryRow(ctx, `
		SELECT d.id, d.name, COALESCE(d.address, ''), COALESCE(d.phone, ''),
			i.id, COALESCE(i.invoice_no, ''), i.period_year, i.period_month, i.total_amount::float8,
			COALESCE((SELECT SUM(ap.total_amount) FROM payments ap WHERE ap.invoice_id = i.id AND ap.status = 'active'), 0)::float8
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		JOIN contracts c ON c.id = i.contract_id
		JOIN rooms rm ON rm.id = c.room_id
		JOIN dormitories d ON d.id = rm.dormitory_id
		WHERE p.id = $1
	`, id).Scan(
		&receipt.Dormitory.ID, &receipt.Dormitory.Name, &receipt.Dormitory.Address, &receipt.Dormitory.Phone,
		&receipt.Invoice.ID, &receipt.Invoice.InvoiceNo, &receipt.Invoice.PeriodYear, &receipt.Invoice.PeriodMonth, &receipt.Invoice.TotalAmount,
		&receipt.Invoice.PaidAmount,
	)
	if err != nil {
		return paymentdomain.Receipt{}, err
	}
	return receipt, nil
}
