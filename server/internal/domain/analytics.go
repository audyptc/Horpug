package domain

import "context"

type MonthlyRevenue struct {
	Month   string  `json:"month"`
	Revenue float64 `json:"revenue"`
}

type MonthlyCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type BillStatusCount struct {
	Unpaid  int `json:"unpaid"`
	Paid    int `json:"paid"`
	Overdue int `json:"overdue"`
}

type AnalyticsSummary struct {
	MonthlyRevenue  []MonthlyRevenue `json:"monthly_revenue"`
	MonthlyTenants  []MonthlyCount   `json:"monthly_tenants"`
	BillStatus      BillStatusCount  `json:"bill_status"`
	TotalRevenue    float64          `json:"total_revenue"`
	AvgMonthly      float64          `json:"avg_monthly"`
}

type AnalyticsRepository interface {
	Summary(ctx context.Context, months int) (*AnalyticsSummary, error)
}
