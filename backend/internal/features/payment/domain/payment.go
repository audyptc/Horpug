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
	Items         []PaymentItem `json:"items"`
	CreatedBy     *uuid.UUID    `json:"created_by,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}
