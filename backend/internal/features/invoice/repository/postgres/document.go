package postgres

import (
	"context"

	invoicedomain "apihorpug/internal/features/invoice/domain"

	"github.com/google/uuid"
)

// GetDocument loads an invoice with its issuing dormitory and payments for
// printing. Access follows GetByID: a missing invoice and one outside the
// requester's dormitories both surface as ErrInvoiceNotFound.
func (r *Repository) GetDocument(ctx context.Context, id, requesterID uuid.UUID) (invoicedomain.Document, error) {
	invoice, err := r.GetByID(ctx, id, requesterID)
	if err != nil {
		return invoicedomain.Document{}, err
	}

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
		SELECT id, payment_date, total_amount, COALESCE(note, '')
		FROM payments
		WHERE invoice_id = $1
		ORDER BY payment_date ASC, created_at ASC
	`, id)
	if err != nil {
		return invoicedomain.Document{}, err
	}
	for rows.Next() {
		var p invoicedomain.DocumentPayment
		if err := rows.Scan(&p.ID, &p.PaymentDate, &p.TotalAmount, &p.Note); err != nil {
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
