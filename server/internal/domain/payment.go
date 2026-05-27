package domain

import (
	"context"
	"time"
)

type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodQR       PaymentMethod = "qr"
)

type Payment struct {
	ID          string        `json:"id"`
	BillID      string        `json:"bill_id"`
	Amount      float64       `json:"amount"`
	Method      PaymentMethod `json:"method"`
	PaymentDate time.Time     `json:"payment_date"`
	Note        string        `json:"note"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type PaymentDetail struct {
	Payment
	RoomNumber      string  `json:"room_number"`
	TenantFirstName string  `json:"tenant_first_name"`
	TenantLastName  string  `json:"tenant_last_name"`
	BillingMonth    string  `json:"billing_month"`
	BillTotal       float64 `json:"bill_total"`
}

type CreatePaymentRequest struct {
	BillID      string        `json:"bill_id"`
	Amount      float64       `json:"amount"`
	Method      PaymentMethod `json:"method"`
	PaymentDate time.Time     `json:"payment_date"`
	Note        string        `json:"note"`
}

type UpdatePaymentRequest struct {
	Amount      float64       `json:"amount"`
	Method      PaymentMethod `json:"method"`
	PaymentDate time.Time     `json:"payment_date"`
	Note        string        `json:"note"`
}

type PaymentRepository interface {
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindDetailByID(ctx context.Context, id string) (*PaymentDetail, error)
	List(ctx context.Context, limit, offset int) ([]*PaymentDetail, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, p *Payment) error
	Update(ctx context.Context, p *Payment) error
	Delete(ctx context.Context, id string) error
}
