package domain

import (
	"context"
	"time"
)

type BillStatus string

const (
	BillStatusUnpaid  BillStatus = "unpaid"
	BillStatusPaid    BillStatus = "paid"
	BillStatusOverdue BillStatus = "overdue"
)

type Bill struct {
	ID             string     `json:"id"`
	ContractID     string     `json:"contract_id"`
	BillingMonth   time.Time  `json:"billing_month"`
	RentAmount     float64    `json:"rent_amount"`
	ElectricAmount float64    `json:"electric_amount"`
	WaterAmount    float64    `json:"water_amount"`
	OtherAmount    float64    `json:"other_amount"`
	OtherNote      string     `json:"other_note"`
	TotalAmount    float64    `json:"total_amount"`
	Status         BillStatus `json:"status"`
	DueDate        *time.Time `json:"due_date"`
	PaidAt         *time.Time `json:"paid_at"`
	Note           string     `json:"note"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type BillDetail struct {
	Bill
	TenantFirstName string `json:"tenant_first_name"`
	TenantLastName  string `json:"tenant_last_name"`
	RoomNumber      string `json:"room_number"`
}

type CreateBillRequest struct {
	ContractID     string     `json:"contract_id"`
	BillingMonth   time.Time  `json:"billing_month"`
	RentAmount     float64    `json:"rent_amount"`
	ElectricAmount float64    `json:"electric_amount"`
	WaterAmount    float64    `json:"water_amount"`
	OtherAmount    float64    `json:"other_amount"`
	OtherNote      string     `json:"other_note"`
	DueDate        *time.Time `json:"due_date"`
	Note           string     `json:"note"`
}

type UpdateBillRequest struct {
	RentAmount     float64    `json:"rent_amount"`
	ElectricAmount float64    `json:"electric_amount"`
	WaterAmount    float64    `json:"water_amount"`
	OtherAmount    float64    `json:"other_amount"`
	OtherNote      string     `json:"other_note"`
	Status         BillStatus `json:"status"`
	DueDate        *time.Time `json:"due_date"`
	Note           string     `json:"note"`
}

type BillRepository interface {
	FindByID(ctx context.Context, id string) (*Bill, error)
	FindDetailByID(ctx context.Context, id string) (*BillDetail, error)
	List(ctx context.Context, limit, offset int) ([]*BillDetail, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, b *Bill) error
	Update(ctx context.Context, b *Bill) error
	Delete(ctx context.Context, id string) error
}
