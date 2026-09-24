package usecase

import (
	"context"
	"log"
	"math"

	invoicedomain "apihorpug/internal/features/invoice/domain"
	"apihorpug/internal/platform/promptpay"

	"github.com/google/uuid"
)

// DocumentRepository is the part of the repository the printable document needs.
type DocumentRepository interface {
	GetDocument(ctx context.Context, id, requesterID uuid.UUID) (invoicedomain.Document, error)
	GetDocumentForTenant(ctx context.Context, id, tenantID uuid.UUID) (invoicedomain.Document, error)
}

// GetDocument returns the invoice as a printable invoice/receipt, adding the
// outstanding amount and a PromptPay QR payload for it when one applies.
func (s *Service) GetDocument(ctx context.Context, id, requesterID uuid.UUID) (invoicedomain.Document, error) {
	doc, err := s.repo.GetDocument(ctx, id, requesterID)
	if err != nil {
		return invoicedomain.Document{}, err
	}
	return finishDocument(doc), nil
}

// GetDocumentForTenant is GetDocument for the tenant self-service pages,
// limited to the tenant's own invoices.
func (s *Service) GetDocumentForTenant(ctx context.Context, id, tenantID uuid.UUID) (invoicedomain.Document, error) {
	doc, err := s.repo.GetDocumentForTenant(ctx, id, tenantID)
	if err != nil {
		return invoicedomain.Document{}, err
	}
	return finishDocument(doc), nil
}

// finishDocument works out the outstanding amount and the PromptPay payload.
func finishDocument(doc invoicedomain.Document) invoicedomain.Document {
	// Amounts are NUMERIC(10,2) read into float64; round so float noise never
	// shows up as a satang owed.
	doc.Outstanding = math.Max(0, math.Round((doc.Invoice.TotalAmount-doc.PaidAmount)*100)/100)

	owing := doc.Invoice.Status == invoicedomain.InvoiceStatusUnpaid || doc.Invoice.Status == invoicedomain.InvoiceStatusOverdue
	if owing && doc.Outstanding > 0 && doc.Dormitory.PromptPayID != "" {
		payload, err := promptpay.Payload(doc.Dormitory.PromptPayID, doc.Outstanding)
		if err != nil {
			// Stored IDs are validated on save, so this means bad legacy
			// data; print the document without a QR rather than failing.
			log.Printf("invalid promptpay id on dormitory %s: %v", doc.Dormitory.ID, err)
		} else {
			doc.PromptPayPayload = payload
		}
	}

	return doc
}
