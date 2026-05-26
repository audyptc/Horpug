package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"
)

type AnalyticsRepo struct {
	db *database.DB
}

func NewAnalyticsRepo(db *database.DB) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

func (r *AnalyticsRepo) Summary(ctx context.Context, months int) (*domain.AnalyticsSummary, error) {
	s := &domain.AnalyticsSummary{}

	// Monthly revenue (paid bills) for last N months
	rows, err := r.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT
			TO_CHAR(gs.month, 'YYYY-MM') AS month,
			COALESCE(SUM(b.total_amount), 0) AS revenue
		FROM generate_series(
			DATE_TRUNC('month', NOW()) - INTERVAL '%d months',
			DATE_TRUNC('month', NOW()),
			INTERVAL '1 month'
		) AS gs(month)
		LEFT JOIN bills b
			ON DATE_TRUNC('month', b.billing_month) = gs.month
			AND b.status = 'paid'
		GROUP BY gs.month
		ORDER BY gs.month
	`, months-1))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var mr domain.MonthlyRevenue
		if err := rows.Scan(&mr.Month, &mr.Revenue); err != nil {
			return nil, err
		}
		s.MonthlyRevenue = append(s.MonthlyRevenue, mr)
		s.TotalRevenue += mr.Revenue
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(s.MonthlyRevenue) > 0 {
		s.AvgMonthly = s.TotalRevenue / float64(len(s.MonthlyRevenue))
	}

	// Monthly new tenants for last N months
	trows, err := r.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT
			TO_CHAR(gs.month, 'YYYY-MM') AS month,
			COUNT(t.id) AS cnt
		FROM generate_series(
			DATE_TRUNC('month', NOW()) - INTERVAL '%d months',
			DATE_TRUNC('month', NOW()),
			INTERVAL '1 month'
		) AS gs(month)
		LEFT JOIN tenants t
			ON DATE_TRUNC('month', t.created_at) = gs.month
		GROUP BY gs.month
		ORDER BY gs.month
	`, months-1))
	if err != nil {
		return nil, err
	}
	defer trows.Close()
	for trows.Next() {
		var mc domain.MonthlyCount
		if err := trows.Scan(&mc.Month, &mc.Count); err != nil {
			return nil, err
		}
		s.MonthlyTenants = append(s.MonthlyTenants, mc)
	}
	if err := trows.Err(); err != nil {
		return nil, err
	}

	// Bill status counts
	err = r.db.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'unpaid')  AS unpaid,
			COUNT(*) FILTER (WHERE status = 'paid')    AS paid,
			COUNT(*) FILTER (WHERE status = 'overdue') AS overdue
		FROM bills
	`).Scan(&s.BillStatus.Unpaid, &s.BillStatus.Paid, &s.BillStatus.Overdue)
	if err != nil {
		return nil, err
	}

	return s, nil
}
