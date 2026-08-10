package repository

import (
	"context"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/feature/dashboard/domain"
)

type DashboardRepo struct {
	db *database.DB
}

func NewDashboardRepo(db *database.DB) *DashboardRepo {
	return &DashboardRepo{db: db}
}

func (r *DashboardRepo) Summary(ctx context.Context) (*domain.DashboardSummary, error) {
	s := &domain.DashboardSummary{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM rooms)                                                                    AS total_rooms,
			(SELECT COUNT(*) FROM rooms WHERE status = 'occupied')                                         AS occupied_rooms,
			(SELECT COUNT(*) FROM rooms WHERE status = 'available')                                        AS available_rooms,
			(SELECT COUNT(*) FROM rooms WHERE status = 'maintenance')                                      AS maintenance_rooms,
			(SELECT COUNT(*) FROM tenants)                                                                  AS total_tenants,
			(SELECT COUNT(*) FROM contracts WHERE status = 'active')                                       AS active_contracts,
			(SELECT COUNT(*) FROM contracts
			   WHERE status = 'active'
			     AND end_date IS NOT NULL
			     AND end_date BETWEEN NOW() AND NOW() + INTERVAL '30 days')                                AS expiring_contracts,
			(SELECT COUNT(*) FROM bills WHERE status = 'unpaid')                                           AS unpaid_bills,
			(SELECT COUNT(*) FROM bills WHERE status = 'overdue')                                          AS overdue_bills,
			(SELECT COALESCE(SUM(total_amount), 0) FROM bills WHERE status = 'unpaid')                    AS unpaid_amount,
			(SELECT COALESCE(SUM(total_amount), 0) FROM bills WHERE status = 'overdue')                   AS overdue_amount,
			(SELECT COALESCE(SUM(total_amount), 0) FROM bills
			   WHERE status = 'paid'
			     AND DATE_TRUNC('month', billing_month) = DATE_TRUNC('month', NOW()))                      AS revenue_this_month
	`).Scan(
		&s.TotalRooms, &s.OccupiedRooms, &s.AvailableRooms, &s.MaintenanceRooms,
		&s.TotalTenants, &s.ActiveContracts, &s.ExpiringContracts,
		&s.UnpaidBills, &s.OverdueBills,
		&s.UnpaidAmount, &s.OverdueAmount, &s.RevenueThisMonth,
	)
	if err != nil {
		return nil, err
	}
	if s.TotalRooms > 0 {
		s.OccupancyRate = float64(s.OccupiedRooms) / float64(s.TotalRooms) * 100
	}
	return s, nil
}
