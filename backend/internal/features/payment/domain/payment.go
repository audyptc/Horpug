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

type Payment struct {
	ID            uuid.UUID     `json:"id"`
	InvoiceID     uuid.UUID     `json:"invoice_id"`
	TenantID      uuid.UUID     `json:"tenant_id,omitempty"`
	TenantName    string        `json:"tenant_name,omitempty"`
	RoomID        uuid.UUID     `json:"room_id,omitempty"`
	RoomNumber    string        `json:"room_number,omitempty"`
	DormitoryID   uuid.UUID     `json:"dormitory_id,omitempty"`
	DormitoryName string        `json:"dormitory_name,omitempty"`
	Amount        float64       `json:"amount"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	PaymentDate   time.Time     `json:"payment_date"`
	ReferenceNo   string        `json:"reference_no"`
	Note          string        `json:"note"`
	CreatedBy     *uuid.UUID    `json:"created_by,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}
