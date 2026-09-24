package domain

import (
	"time"

	"github.com/google/uuid"
)

// ItemType classifies a line of a move-out settlement.
type ItemType string

const (
	// ItemInvoice is the unpaid balance of an existing invoice; it is settled
	// from the deposit as a payment with its own receipt.
	ItemInvoice ItemType = "invoice"
	// ItemRent is rent for the days stayed in the move-out month, when that
	// month hasn't been invoiced.
	ItemRent ItemType = "rent"
	// ItemRentCredit gives back rent already invoiced for days after the
	// move-out date (a negative amount).
	ItemRentCredit ItemType = "rent_credit"
	// ItemElectricity and ItemWater are meter readings not yet on any invoice.
	ItemElectricity ItemType = "electricity"
	ItemWater       ItemType = "water"
	// ItemOther is a deduction entered by staff (cleaning, repairs, keys...).
	ItemOther ItemType = "other"
)

// Item is one line of a settlement. Positive amounts are deducted from the
// deposit; a negative amount (rent credit) is added back.
type Item struct {
	ItemType    ItemType   `json:"item_type"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount"`
	ReferenceID *uuid.UUID `json:"reference_id,omitempty"`
}

// Settlement is the deposit reckoning for a contract ending on MoveOutDate.
// RefundAmount is what the tenant gets back; AmountDue is what they still owe
// when the deductions exceed the deposit. At most one of them is non-zero.
type Settlement struct {
	ContractID      uuid.UUID `json:"contract_id"`
	TenantName      string    `json:"tenant_name"`
	RoomID          uuid.UUID `json:"room_id"`
	RoomNumber      string    `json:"room_number"`
	DormitoryID     uuid.UUID `json:"dormitory_id"`
	DormitoryName   string    `json:"dormitory_name"`
	StartDate       time.Time `json:"start_date"`
	MoveOutDate     time.Time `json:"move_out_date"`
	RentPrice       float64   `json:"rent_price"`
	Deposit         float64   `json:"deposit"`
	DaysStayed      int       `json:"days_stayed"`
	DaysInMonth     int       `json:"days_in_month"`
	Items           []Item    `json:"items"`
	TotalDeductions float64   `json:"total_deductions"`
	RefundAmount    float64   `json:"refund_amount"`
	AmountDue       float64   `json:"amount_due"`
}

// Preview is a settlement as it would be confirmed, plus the dormitory's
// ready-made deductions to pick from.
type Preview struct {
	Settlement
	Presets []Preset `json:"presets"`
}

// DepositPayment is a payment recorded against an unpaid invoice from the
// deposit when the move-out was confirmed.
type DepositPayment struct {
	InvoiceID uuid.UUID `json:"invoice_id"`
	PaymentID uuid.UUID `json:"payment_id"`
	ReceiptNo string    `json:"receipt_no"`
	Amount    float64   `json:"amount"`
}

// MoveOut is a confirmed settlement: the contract is terminated and the
// deposit applied.
type MoveOut struct {
	ID uuid.UUID `json:"id"`
	Settlement
	DormitoryAddress string           `json:"dormitory_address"`
	DormitoryPhone   string           `json:"dormitory_phone"`
	Note             string           `json:"note"`
	DepositPayments  []DepositPayment `json:"deposit_payments"`
	CreatedBy        *uuid.UUID       `json:"created_by,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

// Preset is a ready-made deduction a dormitory offers at move-out, e.g.
// "ค่าทำความสะอาด 500".
type Preset struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Amount float64   `json:"amount"`
}
