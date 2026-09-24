package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod string

const (
	PaymentMethodCash       PaymentMethod = "cash"
	PaymentMethodTransfer   PaymentMethod = "transfer"
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodOther      PaymentMethod = "other"
)

func (m PaymentMethod) Valid() bool {
	switch m {
	case PaymentMethodCash, PaymentMethodTransfer, PaymentMethodCreditCard, PaymentMethodOther:
		return true
	}
	return false
}

// PaymentItem is one payment-method line within a Payment, e.g. the "cash
// 3,000" or "transfer 2,000" portion of a single receipt split across
// multiple methods.
type PaymentItem struct {
	ID            uuid.UUID     `json:"id"`
	PaymentID     uuid.UUID     `json:"payment_id"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Amount        float64       `json:"amount"`
	ReferenceNo   string        `json:"reference_no"`
	CreatedAt     time.Time     `json:"created_at"`
}

// InvoiceStatusChange reports the invoice status flip a payment change caused
// (unpaid -> paid when payments reach the total, paid -> unpaid when they fall
// short). Both fields are empty when the status was left alone.
type InvoiceStatusChange struct {
	From string
	To   string
}

func (c InvoiceStatusChange) Changed() bool {
	return c.To != ""
}

type PaymentStatus string

const (
	PaymentStatusActive PaymentStatus = "active"
	// PaymentStatusVoided marks a cancelled receipt. It is kept (and its
	// number never reused) but no longer counts towards the invoice.
	PaymentStatusVoided PaymentStatus = "voided"
)

// Payment.ReceiptNo is the receipt number, RC<year>-<seq>, counted per
// dormitory and restarting each year.
type Payment struct {
	ID            uuid.UUID     `json:"id"`
	InvoiceID     uuid.UUID     `json:"invoice_id"`
	TenantID      uuid.UUID     `json:"tenant_id,omitempty"`
	TenantName    string        `json:"tenant_name,omitempty"`
	RoomID        uuid.UUID     `json:"room_id,omitempty"`
	RoomNumber    string        `json:"room_number,omitempty"`
	DormitoryID   uuid.UUID     `json:"dormitory_id,omitempty"`
	DormitoryName string        `json:"dormitory_name,omitempty"`
	TotalAmount   float64       `json:"total_amount"`
	PaymentDate   time.Time     `json:"payment_date"`
	Note          string        `json:"note"`
	ReceiptNo     string        `json:"receipt_no"`
	Status        PaymentStatus `json:"status"`
	VoidedAt      *time.Time    `json:"voided_at,omitempty"`
	VoidReason    string        `json:"void_reason,omitempty"`
	Items         []PaymentItem `json:"items"`
	CreatedBy     *uuid.UUID    `json:"created_by,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}
