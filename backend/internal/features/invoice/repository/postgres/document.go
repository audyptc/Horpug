package postgres

import (
	"context"
	"errors"

	invoicedomain "apihorpug/internal/features/invoice/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetDocument loads an invoice with its issuing dormitory and payments for
// printing. Access follows GetByID: a missing invoice and one outside the
// requester's dormitories both surface as ErrInvoiceNotFound.
func (r *Repository) GetDocument(ctx context.Context, id, requesterID uuid.UUID) (invoicedomain.Document, error) {
	invoice, err := r.GetByID(ctx, id, requesterID)
	if err != nil {
		return invoicedomain.Document{}, err
	}
	return r.buildDocument(ctx, invoice)
}

// GetDocumentForTenant loads the document for the tenant self-service pages:
// only an invoice on one of the tenant's own contracts, cancelled ones
// excluded. Anything else surfaces as ErrInvoiceNotFound.
func (r *Repository) GetDocumentForTenant(ctx context.Context, id, tenantID uuid.UUID) (invoicedomain.Document, error) {
	var owned bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM invoices i JOIN contracts c ON c.id = i.contract_id
			WHERE i.id = $1 AND c.tenant_id = $2 AND i.status <> 'cancelled'
		)
	`, id, tenantID).Scan(&owned); err != nil {
		return invoicedomain.Document{}, err
	}
	if !owned {
		return invoicedomain.Document{}, invoicedomain.ErrInvoiceNotFound
	}

	invoice, err := r.loadInvoiceByID(ctx, id)
	if err != nil {
		return invoicedomain.Document{}, err
	}
	if invoice.Items, err = r.loadInvoiceItems(ctx, id); err != nil {
		return invoicedomain.Document{}, err
	}
	return r.buildDocument(ctx, invoice)
}

// buildDocument adds the issuing dormitory and the active payments to an
// invoice the caller has already been allowed to see.
func (r *Repository) buildDocument(ctx context.Context, invoice invoicedomain.Invoice) (invoicedomain.Document, error) {
	id := invoice.ID
	doc := invoicedomain.Document{Invoice: invoice, Payments: make([]invoicedomain.DocumentPayment, 0)}
	if err := r.db.QueryRow(ctx, `
		SELECT d.id, d.name, COALESCE(d.address, ''), COALESCE(d.phone, ''), d.promptpay_id
		FROM invoices i
		JOIN contracts c ON c.id = i.contract_id
		JOIN rooms rm ON rm.id = c.room_id
		JOIN dormitories d ON d.id = rm.dormitory_id
		WHERE i.id = $1
	`, id).Scan(&doc.Dormitory.ID, &doc.Dormitory.Name, &doc.Dormitory.Address, &doc.Dormitory.Phone, &doc.Dormitory.PromptPayID); err != nil {
		return invoicedomain.Document{}, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, COALESCE(receipt_no, ''), payment_date, total_amount, COALESCE(note, '')
		FROM payments
		WHERE invoice_id = $1 AND status = 'active'
		ORDER BY payment_date ASC, created_at ASC
	`, id)
	if err != nil {
		return invoicedomain.Document{}, err
	}
	for rows.Next() {
		var p invoicedomain.DocumentPayment
		if err := rows.Scan(&p.ID, &p.ReceiptNo, &p.PaymentDate, &p.TotalAmount, &p.Note); err != nil {
			rows.Close()
			return invoicedomain.Document{}, err
		}
		p.Items = make([]invoicedomain.DocumentPaymentItem, 0)
		doc.Payments = append(doc.Payments, p)
		doc.PaidAmount += p.TotalAmount
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return invoicedomain.Document{}, err
	}

	if len(doc.Payments) == 0 {
		return doc, nil
	}

	index := make(map[uuid.UUID]int, len(doc.Payments))
	ids := make([]uuid.UUID, len(doc.Payments))
	for i, p := range doc.Payments {
		index[p.ID] = i
		ids[i] = p.ID
	}

	itemRows, err := r.db.Query(ctx, `
		SELECT payment_id, payment_method, amount, COALESCE(reference_no, '')
		FROM payment_items
		WHERE payment_id = ANY($1)
		ORDER BY created_at ASC
	`, ids)
	if err != nil {
		return invoicedomain.Document{}, err
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var paymentID uuid.UUID
		var item invoicedomain.DocumentPaymentItem
		if err := itemRows.Scan(&paymentID, &item.PaymentMethod, &item.Amount, &item.ReferenceNo); err != nil {
			return invoicedomain.Document{}, err
		}
		p := &doc.Payments[index[paymentID]]
		p.Items = append(p.Items, item)
	}
	if err := itemRows.Err(); err != nil {
		return invoicedomain.Document{}, err
	}

	return doc, nil
}

// GetQRInfo loads an invoice's status, balance and PromptPay account for the
// public QR image. There is no requester: the caller has already checked the
// signed link.
func (r *Repository) GetQRInfo(ctx context.Context, id uuid.UUID) (invoicedomain.QRInfo, error) {
	var info invoicedomain.QRInfo
	err := r.db.QueryRow(ctx, `
		SELECT i.status, d.promptpay_id, i.total_amount::float8,
			COALESCE((SELECT SUM(p.total_amount) FROM payments p WHERE p.invoice_id = i.id AND p.status = 'active'), 0)::float8
		FROM invoices i
		JOIN contracts c ON c.id = i.contract_id
		JOIN rooms rm ON rm.id = c.room_id
		JOIN dormitories d ON d.id = rm.dormitory_id
		WHERE i.id = $1
	`, id).Scan(&info.Status, &info.PromptPayID, &info.TotalAmount, &info.PaidAmount)
	if errors.Is(err, pgx.ErrNoRows) {
		return invoicedomain.QRInfo{}, invoicedomain.ErrInvoiceNotFound
	}
	return info, err
}
