package domain

// RoomStats counts rooms by status.
type RoomStats struct {
	Total       int64 `json:"total"`
	Available   int64 `json:"available"`
	Occupied    int64 `json:"occupied"`
	Maintenance int64 `json:"maintenance"`
}

// ContractStats counts contracts overall and those currently active, plus the
// active ones that need a decision about their end date.
type ContractStats struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
	// ExpiringSoon counts active contracts ending within the expiry window
	// (today included).
	ExpiringSoon int64 `json:"expiring_soon"`
	// PastEnd counts contracts still marked active although their end date has
	// already passed: nobody has renewed or terminated them.
	PastEnd int64 `json:"past_end"`
}

// InvoiceStats covers invoices still awaiting payment (unpaid or overdue).
type InvoiceStats struct {
	OutstandingCount  int64   `json:"outstanding_count"`
	OutstandingAmount float64 `json:"outstanding_amount"`
	// Overdue is the part of the outstanding total that is past its due date.
	// Judged from the due date rather than the stored status, which the
	// hourly sweeper (invoice usecase RunOverdueSweeper) can lag by up to an
	// hour.
	OverdueCount  int64   `json:"overdue_count"`
	OverdueAmount float64 `json:"overdue_amount"`
}

// RepairStats counts repair requests that still need attention.
type RepairStats struct {
	Pending    int64 `json:"pending"`
	InProgress int64 `json:"in_progress"`
}

// PaymentStats totals the payments received in the current month.
type PaymentStats struct {
	ReceivedThisMonth float64 `json:"received_this_month"`
}

// ExpenseStats totals the expenses dated in the current month.
type ExpenseStats struct {
	ThisMonth float64 `json:"this_month"`
}

// ReadingStats counts occupied rooms (an active contract that has started)
// with no meter reading yet in the current month. Invoices are built from the
// readings of their month, so a room missing here bills without that charge.
type ReadingStats struct {
	Missing int64 `json:"missing"`
}

// Period is the calendar month the "this month" figures refer to.
type Period struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

// Summary is the dashboard's headline numbers, counted over the dormitories
// the requester manages. A section is nil when the requester's role has no
// read access to the menu it comes from, so the dashboard never reveals
// figures the matching list page would refuse to show.
type Summary struct {
	Period              Period         `json:"period"`
	Rooms               *RoomStats     `json:"rooms"`
	Contracts           *ContractStats `json:"contracts"`
	Invoices            *InvoiceStats  `json:"invoices"`
	Repairs             *RepairStats   `json:"repairs"`
	Payments            *PaymentStats  `json:"payments"`
	Expenses            *ExpenseStats  `json:"expenses"`
	ElectricityReadings *ReadingStats  `json:"electricity_readings"`
	WaterReadings       *ReadingStats  `json:"water_readings"`
}
