package usecase

import (
	"context"
	"testing"

	invoicedomain "apihorpug/internal/features/invoice/domain"

	"github.com/google/uuid"
)

type fakeDocumentRepo struct {
	Repository
	doc invoicedomain.Document
}

func (f *fakeDocumentRepo) GetDocument(context.Context, uuid.UUID, uuid.UUID) (invoicedomain.Document, error) {
	return f.doc, nil
}

func TestGetDocumentPromptPay(t *testing.T) {
	cases := []struct {
		name            string
		status          invoicedomain.InvoiceStatus
		total, paid     float64
		promptPayID     string
		wantOutstanding float64
		wantQR          bool
	}{
		{"unpaid with promptpay", invoicedomain.InvoiceStatusUnpaid, 4200, 0, "0812345678", 4200, true},
		{"overdue part paid", invoicedomain.InvoiceStatusOverdue, 4200, 1200, "0812345678", 3000, true},
		{"no promptpay on dormitory", invoicedomain.InvoiceStatusUnpaid, 4200, 0, "", 4200, false},
		{"paid", invoicedomain.InvoiceStatusPaid, 4200, 4200, "0812345678", 0, false},
		{"cancelled", invoicedomain.InvoiceStatusCancelled, 4200, 0, "0812345678", 4200, false},
		{"float noise rounds away", invoicedomain.InvoiceStatusUnpaid, 0.3, 0.1 + 0.2, "0812345678", 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeDocumentRepo{doc: invoicedomain.Document{
				Invoice:    invoicedomain.Invoice{Status: tc.status, TotalAmount: tc.total},
				Dormitory:  invoicedomain.DocumentDormitory{PromptPayID: tc.promptPayID},
				PaidAmount: tc.paid,
			}}
			doc, err := New(repo, nil, nil).GetDocument(context.Background(), uuid.New(), uuid.New())
			if err != nil {
				t.Fatalf("GetDocument() error = %v", err)
			}
			if doc.Outstanding != tc.wantOutstanding {
				t.Errorf("Outstanding = %v, want %v", doc.Outstanding, tc.wantOutstanding)
			}
			if (doc.PromptPayPayload != "") != tc.wantQR {
				t.Errorf("PromptPayPayload = %q, want QR: %v", doc.PromptPayPayload, tc.wantQR)
			}
		})
	}
}
