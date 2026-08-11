package repository

import (
	"context"

	"apigofiberhorpug/internal/features/dashboard/domain"
	"apigofiberhorpug/internal/platform/database"
)

type DashboardRepo struct {
	db *database.DB
}

func NewDashboardRepo(db *database.DB) *DashboardRepo {
	return &DashboardRepo{db: db}
}

func (r *DashboardRepo) Summary(ctx context.Context, dormitoryID string) (*domain.DashboardSummary, error) {
	s := &domain.DashboardSummary{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM rooms WHERE dormitory_id = $1)                                            AS total_rooms,
			(SELECT COUNT(*) FROM rooms WHERE dormitory_id = $1 AND status = 'occupied')                    AS occupied_rooms,
			(SELECT COUNT(*) FROM rooms WHERE dormitory_id = $1 AND status = 'available')                   AS available_rooms,
			(SELECT COUNT(*) FROM rooms WHERE dormitory_id = $1 AND status = 'maintenance')                 AS maintenance_rooms,
			(SELECT COUNT(*) FROM tenants WHERE dormitory_id = $1)                                          AS total_tenants,
			(SELECT COUNT(*) FROM contracts c JOIN rooms r ON r.id = c.room_id
			   WHERE r.dormitory_id = $1 AND c.status = 'active')                                           AS active_contracts,
			(SELECT COUNT(*) FROM contracts c JOIN rooms r ON r.id = c.room_id
			   WHERE r.dormitory_id = $1
			     AND c.status = 'active'
			     AND c.end_date IS NOT NULL
			     AND c.end_date BETWEEN NOW() AND NOW() + INTERVAL '30 days')                                AS expiring_contracts,
			(SELECT COUNT(*) FROM bills b JOIN contracts c ON c.id = b.contract_id JOIN rooms r ON r.id = c.room_id
			   WHERE r.dormitory_id = $1 AND b.status = 'unpaid')                                           AS unpaid_bills,
			(SELECT COUNT(*) FROM bills b JOIN contracts c ON c.id = b.contract_id JOIN rooms r ON r.id = c.room_id
			   WHERE r.dormitory_id = $1 AND b.status = 'overdue')                                          AS overdue_bills,
			(SELECT COALESCE(SUM(b.total_amount), 0) FROM bills b JOIN contracts c ON c.id = b.contract_id JOIN rooms r ON r.id = c.room_id
			   WHERE r.dormitory_id = $1 AND b.status = 'unpaid')                                           AS unpaid_amount,
			(SELECT COALESCE(SUM(b.total_amount), 0) FROM bills b JOIN contracts c ON c.id = b.contract_id JOIN rooms r ON r.id = c.room_id
			   WHERE r.dormitory_id = $1 AND b.status = 'overdue')                                          AS overdue_amount,
			(SELECT COALESCE(SUM(b.total_amount), 0) FROM bills b JOIN contracts c ON c.id = b.contract_id JOIN rooms r ON r.id = c.room_id
			   WHERE r.dormitory_id = $1
			     AND b.status = 'paid'
			     AND DATE_TRUNC('month', b.billing_month) = DATE_TRUNC('month', NOW()))                      AS revenue_this_month
	`, dormitoryID).Scan(
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
