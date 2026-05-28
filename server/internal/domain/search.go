package domain

import "time"

type SearchBill struct {
	ID              string     `json:"id"`
	BillingMonth    time.Time  `json:"billing_month"`
	TotalAmount     float64    `json:"total_amount"`
	Status          BillStatus `json:"status"`
	TenantFirstName string     `json:"tenant_first_name"`
	TenantLastName  string     `json:"tenant_last_name"`
	RoomNumber      string     `json:"room_number"`
}

type SearchContract struct {
	ID              string         `json:"id"`
	StartDate       time.Time      `json:"start_date"`
	EndDate         *time.Time     `json:"end_date"`
	Status          ContractStatus `json:"status"`
	TenantFirstName string         `json:"tenant_first_name"`
	TenantLastName  string         `json:"tenant_last_name"`
	RoomNumber      string         `json:"room_number"`
}

type SearchResults struct {
	Tenants   []*Tenant         `json:"tenants"`
	Rooms     []*Room           `json:"rooms"`
	Bills     []*SearchBill     `json:"bills"`
	Contracts []*SearchContract `json:"contracts"`
}
