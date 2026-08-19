package domain

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceStatusUnpaid    InvoiceStatus = "unpaid"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

func (s InvoiceStatus) Valid() bool {
	switch s {
	case InvoiceStatusUnpaid, InvoiceStatusPaid, InvoiceStatusOverdue, InvoiceStatusCancelled:
		return true
	}
	return false
}

type InvoiceItemType string

const (
	InvoiceItemTypeRent        InvoiceItemType = "rent"
	InvoiceItemTypeElectricity InvoiceItemType = "electricity"
	InvoiceItemTypeWater       InvoiceItemType = "water"
	InvoiceItemTypeOther       InvoiceItemType = "other"
)

type InvoiceItem struct {
	ID          uuid.UUID       `json:"id"`
	InvoiceID   uuid.UUID       `json:"invoice_id"`
	ItemType    InvoiceItemType `json:"item_type"`
	Description string          `json:"description"`
	ReferenceID *uuid.UUID      `json:"reference_id,omitempty"`
	Amount      float64         `json:"amount"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Invoice struct {
	ID            uuid.UUID     `json:"id"`
	ContractID    uuid.UUID     `json:"contract_id"`
	TenantID      uuid.UUID     `json:"tenant_id,omitempty"`
	TenantName    string        `json:"tenant_name,omitempty"`
	TenantLineID  string        `json:"tenant_line_id,omitempty"`
	RoomID        uuid.UUID     `json:"room_id,omitempty"`
	RoomNumber    string        `json:"room_number,omitempty"`
	DormitoryID   uuid.UUID     `json:"dormitory_id,omitempty"`
	DormitoryName string        `json:"dormitory_name,omitempty"`
	PeriodYear    int           `json:"period_year"`
	PeriodMonth   int           `json:"period_month"`
	IssueDate     time.Time     `json:"issue_date"`
	DueDate       time.Time     `json:"due_date"`
	TotalAmount   float64       `json:"total_amount"`
	Status        InvoiceStatus `json:"status"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	Note          string        `json:"note"`
	Items         []InvoiceItem `json:"items,omitempty"`
	CreatedBy     *uuid.UUID    `json:"created_by,omitempty"`
	UpdatedBy     *uuid.UUID    `json:"updated_by,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
