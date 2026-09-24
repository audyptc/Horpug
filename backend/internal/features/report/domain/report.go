package domain

import (
	"time"

	"github.com/google/uuid"
)

// Filter is what a report is asked for: one month, over the dormitories the
// requester manages or just DormitoryID.
type Filter struct {
	RequesterID uuid.UUID
	DormitoryID *uuid.UUID
	Year        int
	Month       int
}

// Amount is a labelled total, e.g. income by payment method.
type Amount struct {
	Key    string  `json:"key"`
	Amount float64 `json:"amount"`
}

// IncomeSummary is money recorded as received in the month (active payments
// by payment date), including invoices settled from deposits at move-out
// (method "deposit"), which is shown as its own line.
type IncomeSummary struct {
	Count    int      `json:"count"`
	Total    float64  `json:"total"`
	ByMethod []Amount `json:"by_method"`
}

// BillingSummary covers the month's invoices (by billing period, cancelled
// ones left out): what was billed, how much of it has been collected so far,
// and what is still owed.
type BillingSummary struct {
	InvoiceCount int      `json:"invoice_count"`
	Billed       float64  `json:"billed"`
	Collected    float64  `json:"collected"`
	Outstanding  float64  `json:"outstanding"`
	ByItemType   []Amount `json:"by_item_type"`
}

// ExpenseSummary is the month's expenses by expense date.
type ExpenseSummary struct {
	Total      float64  `json:"total"`
	ByCategory []Amount `json:"by_category"`
}

// Arrears is everything still owed on unpaid or overdue invoices of any
// period, as of now.
type Arrears struct {
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}

// Occupancy is the current room count, not a historical figure.
type Occupancy struct {
	Rooms    int `json:"rooms"`
	Occupied int `json:"occupied"`
}

// DormitoryRow is one dormitory's share of the month.
type DormitoryRow struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Income      float64   `json:"income"`
	Expenses    float64   `json:"expenses"`
	Net         float64   `json:"net"`
	Billed      float64   `json:"billed"`
	Outstanding float64   `json:"outstanding"`
}

// MonthlyReport summarises one month over the dormitories the requester
// manages (or the one dormitory asked for).
type MonthlyReport struct {
	Year        int            `json:"year"`
	Month       int            `json:"month"`
	Income      IncomeSummary  `json:"income"`
	Billing     BillingSummary `json:"billing"`
	Expenses    ExpenseSummary `json:"expenses"`
	Net         float64        `json:"net"`
	Arrears     Arrears        `json:"arrears"`
	Occupancy   Occupancy      `json:"occupancy"`
	Dormitories []DormitoryRow `json:"dormitories"`
}

// PaymentLine, InvoiceLine and ExpenseLine are the detail rows exported with
// the report.
type PaymentLine struct {
	ReceiptNo     string
	PaymentDate   time.Time
	DormitoryName string
	RoomNumber    string
	TenantName    string
	PeriodYear    int
	PeriodMonth   int
	Methods       string
	Amount        float64
	Voided        bool
}

type InvoiceLine struct {
	DormitoryName string
	RoomNumber    string
	TenantName    string
	DueDate       time.Time
	Total         float64
	Paid          float64
	Status        string
}

type ExpenseLine struct {
	ExpenseDate   time.Time
	DormitoryName string
	Category      string
	Description   string
	Amount        float64
}

// Details are the rows behind a monthly report.
type Details struct {
	Payments []PaymentLine
	Invoices []InvoiceLine
	Expenses []ExpenseLine
}
