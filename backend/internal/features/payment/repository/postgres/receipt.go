package postgres

import (
	"context"

	paymentdomain "apihorpug/internal/features/payment/domain"

	"github.com/google/uuid"
)

// GetReceipt loads a payment with its dormitory and invoice balance for
// printing. Access follows GetByID.
func (r *Repository) GetReceipt(ctx context.Context, id, requesterID uuid.UUID) (paymentdomain.Receipt, error) {
	payment, err := r.GetByID(ctx, id, requesterID)
	if err != nil {
		return paymentdomain.Receipt{}, err
	}

	receipt := paymentdomain.Receipt{Payment: payment}
	err = r.db.QueryRow(ctx, `
		SELECT d.id, d.name, COALESCE(d.address, ''), COALESCE(d.phone, ''),
			i.id, i.period_year, i.period_month, i.total_amount::float8,
			COALESCE((SELECT SUM(ap.total_amount) FROM payments ap WHERE ap.invoice_id = i.id AND ap.status = 'active'), 0)::float8
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		JOIN contracts c ON c.id = i.contract_id
		JOIN rooms rm ON rm.id = c.room_id
		JOIN dormitories d ON d.id = rm.dormitory_id
		WHERE p.id = $1
	`, id).Scan(
		&receipt.Dormitory.ID, &receipt.Dormitory.Name, &receipt.Dormitory.Address, &receipt.Dormitory.Phone,
		&receipt.Invoice.ID, &receipt.Invoice.PeriodYear, &receipt.Invoice.PeriodMonth, &receipt.Invoice.TotalAmount,
		&receipt.Invoice.PaidAmount,
	)
	if err != nil {
		return paymentdomain.Receipt{}, err
	}
	return receipt, nil
}
