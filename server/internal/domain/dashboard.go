package domain

import "context"

type DashboardSummary struct {
	TotalRooms        int     `json:"total_rooms"`
	OccupiedRooms     int     `json:"occupied_rooms"`
	AvailableRooms    int     `json:"available_rooms"`
	MaintenanceRooms  int     `json:"maintenance_rooms"`
	OccupancyRate     float64 `json:"occupancy_rate"`
	TotalTenants      int     `json:"total_tenants"`
	ActiveContracts   int     `json:"active_contracts"`
	ExpiringContracts int     `json:"expiring_contracts"`
	UnpaidBills       int     `json:"unpaid_bills"`
	OverdueBills      int     `json:"overdue_bills"`
	UnpaidAmount      float64 `json:"unpaid_amount"`
	OverdueAmount     float64 `json:"overdue_amount"`
	RevenueThisMonth  float64 `json:"revenue_this_month"`
}

type DashboardRepository interface {
	Summary(ctx context.Context) (*DashboardSummary, error)
}
