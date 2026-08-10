package domain

import "context"

type IncomeReportRow struct {
	Month          string  `json:"month"`
	RentAmount     float64 `json:"rent_amount"`
	ElectricAmount float64 `json:"electric_amount"`
	WaterAmount    float64 `json:"water_amount"`
	OtherAmount    float64 `json:"other_amount"`
	TotalAmount    float64 `json:"total_amount"`
	BillCount      int     `json:"bill_count"`
}

type IncomeReport struct {
	Rows          []IncomeReportRow `json:"rows"`
	TotalRent     float64           `json:"total_rent"`
	TotalElectric float64           `json:"total_electric"`
	TotalWater    float64           `json:"total_water"`
	TotalOther    float64           `json:"total_other"`
	GrandTotal    float64           `json:"grand_total"`
}

type ExpenseReportRow struct {
	Month       string  `json:"month"`
	Category    string  `json:"category"`
	TotalAmount float64 `json:"total_amount"`
	Count       int     `json:"count"`
}

type ExpenseReport struct {
	Rows       []ExpenseReportRow `json:"rows"`
	GrandTotal float64            `json:"grand_total"`
}

type OccupancyRow struct {
	RoomNumber string  `json:"room_number"`
	Floor      int     `json:"floor"`
	Type       string  `json:"type"`
	Status     string  `json:"status"`
	TenantName string  `json:"tenant_name"`
	RentPrice  float64 `json:"rent_price"`
}

type OccupancyReport struct {
	Rows          []OccupancyRow `json:"rows"`
	Total         int            `json:"total"`
	Occupied      int            `json:"occupied"`
	Available     int            `json:"available"`
	Maintenance   int            `json:"maintenance"`
	OccupancyRate float64        `json:"occupancy_rate"`
}

type ReportRepository interface {
	IncomeReport(ctx context.Context, from, to string) (*IncomeReport, error)
	ExpenseReport(ctx context.Context, from, to string) (*ExpenseReport, error)
	OccupancyReport(ctx context.Context) (*OccupancyReport, error)
}
