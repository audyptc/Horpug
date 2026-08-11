package repository

import (
	"context"

	"apigofiberhorpug/internal/feature/report/domain"
	"apigofiberhorpug/internal/platform/database"
)

type ReportRepo struct {
	db *database.DB
}

func NewReportRepo(db *database.DB) *ReportRepo {
	return &ReportRepo{db: db}
}

func (r *ReportRepo) IncomeReport(ctx context.Context, dormitoryID, from, to string) (*domain.IncomeReport, error) {
	report := &domain.IncomeReport{}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			TO_CHAR(DATE_TRUNC('month', b.billing_month), 'YYYY-MM') AS month,
			COALESCE(SUM(b.rent_amount), 0)     AS rent_amount,
			COALESCE(SUM(b.electric_amount), 0) AS electric_amount,
			COALESCE(SUM(b.water_amount), 0)    AS water_amount,
			COALESCE(SUM(b.other_amount), 0)    AS other_amount,
			COALESCE(SUM(b.total_amount), 0)    AS total_amount,
			COUNT(*) AS bill_count
		FROM bills b
		JOIN contracts c ON c.id = b.contract_id
		JOIN rooms r ON r.id = c.room_id
		WHERE r.dormitory_id = $1
		  AND b.status = 'paid'
		  AND DATE_TRUNC('month', b.billing_month) >= DATE_TRUNC('month', $2::date)
		  AND DATE_TRUNC('month', b.billing_month) <= DATE_TRUNC('month', $3::date)
		GROUP BY DATE_TRUNC('month', b.billing_month)
		ORDER BY DATE_TRUNC('month', b.billing_month)
	`, dormitoryID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row domain.IncomeReportRow
		if err := rows.Scan(&row.Month, &row.RentAmount, &row.ElectricAmount, &row.WaterAmount, &row.OtherAmount, &row.TotalAmount, &row.BillCount); err != nil {
			return nil, err
		}
		report.Rows = append(report.Rows, row)
		report.TotalRent += row.RentAmount
		report.TotalElectric += row.ElectricAmount
		report.TotalWater += row.WaterAmount
		report.TotalOther += row.OtherAmount
		report.GrandTotal += row.TotalAmount
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return report, nil
}

func (r *ReportRepo) ExpenseReport(ctx context.Context, dormitoryID, from, to string) (*domain.ExpenseReport, error) {
	report := &domain.ExpenseReport{}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			TO_CHAR(DATE_TRUNC('month', expense_date), 'YYYY-MM') AS month,
			category,
			COALESCE(SUM(amount), 0) AS total_amount,
			COUNT(*) AS count
		FROM expenses
		WHERE dormitory_id = $1
		  AND DATE_TRUNC('month', expense_date) >= DATE_TRUNC('month', $2::date)
		  AND DATE_TRUNC('month', expense_date) <= DATE_TRUNC('month', $3::date)
		GROUP BY DATE_TRUNC('month', expense_date), category
		ORDER BY DATE_TRUNC('month', expense_date), category
	`, dormitoryID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row domain.ExpenseReportRow
		if err := rows.Scan(&row.Month, &row.Category, &row.TotalAmount, &row.Count); err != nil {
			return nil, err
		}
		report.Rows = append(report.Rows, row)
		report.GrandTotal += row.TotalAmount
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return report, nil
}

func (r *ReportRepo) OccupancyReport(ctx context.Context, dormitoryID string) (*domain.OccupancyReport, error) {
	report := &domain.OccupancyReport{}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT DISTINCT ON (r.id)
			r.room_number,
			r.floor,
			r.type,
			r.status,
			COALESCE(t.first_name || ' ' || t.last_name, '') AS tenant_name,
			r.rent_price
		FROM rooms r
		LEFT JOIN contracts c ON c.room_id = r.id AND c.status = 'active'
		LEFT JOIN tenants t ON t.id = c.tenant_id
		WHERE r.dormitory_id = $1
		ORDER BY r.id, r.floor, r.room_number
	`, dormitoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row domain.OccupancyRow
		if err := rows.Scan(&row.RoomNumber, &row.Floor, &row.Type, &row.Status, &row.TenantName, &row.RentPrice); err != nil {
			return nil, err
		}
		report.Rows = append(report.Rows, row)
		report.Total++
		switch row.Status {
		case "occupied":
			report.Occupied++
		case "available":
			report.Available++
		case "maintenance":
			report.Maintenance++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if report.Total > 0 {
		report.OccupancyRate = float64(report.Occupied) / float64(report.Total) * 100
	}

	return report, nil
}
