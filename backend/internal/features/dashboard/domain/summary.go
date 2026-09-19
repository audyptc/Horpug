package domain

// RoomStats counts rooms by status.
type RoomStats struct {
	Total       int64 `json:"total"`
	Available   int64 `json:"available"`
	Occupied    int64 `json:"occupied"`
	Maintenance int64 `json:"maintenance"`
}

// ContractStats counts contracts overall and those currently active.
type ContractStats struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
}

// InvoiceStats covers invoices still awaiting payment (unpaid or overdue).
type InvoiceStats struct {
	OutstandingCount  int64   `json:"outstanding_count"`
	OutstandingAmount float64 `json:"outstanding_amount"`
}

// RepairStats counts repair requests that still need attention.
type RepairStats struct {
	Pending    int64 `json:"pending"`
	InProgress int64 `json:"in_progress"`
}

// Summary is the dashboard's headline numbers, counted over the dormitories
// the requester manages. A section is nil when the requester's role has no
// read access to the menu it comes from, so the dashboard never reveals
// figures the matching list page would refuse to show.
type Summary struct {
	Rooms     *RoomStats     `json:"rooms"`
	Contracts *ContractStats `json:"contracts"`
	Invoices  *InvoiceStats  `json:"invoices"`
	Repairs   *RepairStats   `json:"repairs"`
}
