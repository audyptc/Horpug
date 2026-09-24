package postgres

import (
	"context"
	"fmt"
	"strings"

	dashboarddomain "apihorpug/internal/features/dashboard/domain"
	dashboardusecase "apihorpug/internal/features/dashboard/usecase"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Menu paths whose read permission gates each section of the summary — the
// same ones that guard the corresponding list endpoints.
const (
	roomsMenuPath          = "/rooms"
	contractsMenuPath      = "/contracts"
	invoicesMenuPath       = "/invoices"
	repairRequestsMenuPath = "/repair-requests"
	paymentsMenuPath       = "/payments"
	expensesMenuPath       = "/expenses"
	metersMenuPath         = "/meters"
	waterMetersMenuPath    = "/water-meters"
)

// scopeTemplate restricts a query to the dormitories the requester manages,
// where {col} is the query's dormitory column. $1 is the requester, $2 their
// role and $3 whether the role is exempt from dormitory scoping. $4 optionally
// narrows to a single dormitory (NULL means all of them); it is ANDed with the
// access check, so it can only ever narrow what the requester may already see
// and a dormitory outside their scope simply counts as zero.
//
// Any further parameters a query needs start at $5.
const scopeTemplate = `(($3 OR {col} IN (
	SELECT dormitory_id FROM user_dormitories WHERE user_id = $1
	UNION
	SELECT dormitory_id FROM role_dormitories WHERE role_id = $2
)) AND ($4::uuid IS NULL OR {col} = $4::uuid))`

func scopeFor(column string) string {
	return strings.ReplaceAll(scopeTemplate, "{col}", column)
}

var (
	// scopeCondition is for queries on a row aliased rm (rooms).
	scopeCondition = scopeFor("rm.dormitory_id")
	// expenseScope is for expenses, which carry their dormitory directly.
	expenseScope = scopeFor("ex.dormitory_id")
)

// GetSummary counts in the database rather than over a fetched page, so the
// figures stay correct however many rows exist.
func (r *Repository) GetSummary(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, window dashboardusecase.Window) (dashboarddomain.Summary, error) {
	summary := dashboarddomain.Summary{
		Period: dashboarddomain.Period{Year: window.Year, Month: window.Month},
	}

	var full bool
	var roleID uuid.UUID
	if err := r.db.QueryRow(ctx, `
		SELECT r.full_dormitory_access, r.id
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, requesterID).Scan(&full, &roleID); err != nil {
		return summary, err
	}

	readable, err := r.readableMenus(ctx, requesterID)
	if err != nil {
		return summary, err
	}

	// Every query takes the scope arguments first; queries that also look at
	// dates append theirs after them, numbered from $5.
	scopeArgs := []any{requesterID, roleID, full, dormitoryID}
	with := func(extra ...any) []any { return append(append([]any{}, scopeArgs...), extra...) }

	if readable[roomsMenuPath] {
		stats := dashboarddomain.RoomStats{}
		if err := r.db.QueryRow(ctx, fmt.Sprintf(`
			SELECT
				COUNT(*),
				COUNT(*) FILTER (WHERE rm.status = 'available'),
				COUNT(*) FILTER (WHERE rm.status = 'occupied'),
				COUNT(*) FILTER (WHERE rm.status = 'maintenance')
			FROM rooms rm
			WHERE %s
		`, scopeCondition), with()...).Scan(&stats.Total, &stats.Available, &stats.Occupied, &stats.Maintenance); err != nil {
			return summary, err
		}
		summary.Rooms = &stats
	}

	if readable[contractsMenuPath] {
		// $5 is today, $6 the last day of the expiry window.
		stats := dashboarddomain.ContractStats{}
		if err := r.db.QueryRow(ctx, fmt.Sprintf(`
			SELECT
				COUNT(*),
				COUNT(*) FILTER (WHERE c.status = 'active'),
				COUNT(*) FILTER (WHERE c.status = 'active' AND c.end_date >= $5::date AND c.end_date <= $6::date),
				COUNT(*) FILTER (WHERE c.status = 'active' AND c.end_date < $5::date)
			FROM contracts c
			JOIN rooms rm ON rm.id = c.room_id
			WHERE %s
		`, scopeCondition), with(window.Today, window.ExpiryLimit)...).Scan(&stats.Total, &stats.Active, &stats.ExpiringSoon, &stats.PastEnd); err != nil {
			return summary, err
		}
		summary.Contracts = &stats
	}

	if readable[invoicesMenuPath] {
		// $5 is today. An unpaid invoice past its due date counts as overdue
		// even though nothing flips its stored status.
		stats := dashboarddomain.InvoiceStats{}
		if err := r.db.QueryRow(ctx, fmt.Sprintf(`
			SELECT
				COUNT(*) FILTER (WHERE i.status IN ('unpaid', 'overdue')),
				COALESCE(SUM(i.total_amount) FILTER (WHERE i.status IN ('unpaid', 'overdue')), 0)::float8,
				COUNT(*) FILTER (WHERE i.status = 'overdue' OR (i.status = 'unpaid' AND i.due_date < $5::date)),
				COALESCE(SUM(i.total_amount) FILTER (WHERE i.status = 'overdue' OR (i.status = 'unpaid' AND i.due_date < $5::date)), 0)::float8
			FROM invoices i
			JOIN contracts c ON c.id = i.contract_id
			JOIN rooms rm ON rm.id = c.room_id
			WHERE %s
		`, scopeCondition), with(window.Today)...).Scan(&stats.OutstandingCount, &stats.OutstandingAmount, &stats.OverdueCount, &stats.OverdueAmount); err != nil {
			return summary, err
		}
		summary.Invoices = &stats
	}

	if readable[repairRequestsMenuPath] {
		stats := dashboarddomain.RepairStats{}
		if err := r.db.QueryRow(ctx, fmt.Sprintf(`
			SELECT
				COUNT(*) FILTER (WHERE rr.status = 'pending'),
				COUNT(*) FILTER (WHERE rr.status = 'in_progress')
			FROM repair_requests rr
			JOIN rooms rm ON rm.id = rr.room_id
			WHERE %s
		`, scopeCondition), with()...).Scan(&stats.Pending, &stats.InProgress); err != nil {
			return summary, err
		}
		summary.Repairs = &stats
	}

	if readable[paymentsMenuPath] {
		// $5..$6 bound the current month, [start, next start).
		stats := dashboarddomain.PaymentStats{}
		if err := r.db.QueryRow(ctx, fmt.Sprintf(`
			SELECT COALESCE(SUM(p.total_amount), 0)::float8
			FROM payments p
			JOIN invoices i ON i.id = p.invoice_id
			JOIN contracts c ON c.id = i.contract_id
			JOIN rooms rm ON rm.id = c.room_id
			WHERE %s AND p.status = 'active' AND p.payment_date >= $5::date AND p.payment_date < $6::date
		`, scopeCondition), with(window.MonthStart, window.NextMonthStart)...).Scan(&stats.ReceivedThisMonth); err != nil {
			return summary, err
		}
		summary.Payments = &stats
	}

	if readable[expensesMenuPath] {
		stats := dashboarddomain.ExpenseStats{}
		if err := r.db.QueryRow(ctx, fmt.Sprintf(`
			SELECT COALESCE(SUM(ex.amount), 0)::float8
			FROM expenses ex
			WHERE %s AND ex.expense_date >= $5::date AND ex.expense_date < $6::date
		`, expenseScope), with(window.MonthStart, window.NextMonthStart)...).Scan(&stats.ThisMonth); err != nil {
			return summary, err
		}
		summary.Expenses = &stats
	}

	if readable[metersMenuPath] {
		missing, err := r.countMissingReadings(ctx, "electricity_meters", with, window)
		if err != nil {
			return summary, err
		}
		summary.ElectricityReadings = &dashboarddomain.ReadingStats{Missing: missing}
	}

	if readable[waterMetersMenuPath] {
		missing, err := r.countMissingReadings(ctx, "water_meters", with, window)
		if err != nil {
			return summary, err
		}
		summary.WaterReadings = &dashboarddomain.ReadingStats{Missing: missing}
	}

	return summary, nil
}

// countMissingReadings counts rooms with a started, active contract that have
// no row in the given meter table dated in the current month. The table name
// is interpolated, so it must only ever be one of the two literals above. At
// most one contract per room is active (unique index), so counting contracts
// counts rooms.
func (r *Repository) countMissingReadings(ctx context.Context, table string, with func(...any) []any, window dashboardusecase.Window) (int64, error) {
	// $5 is today, $6..$7 bound the current month.
	var missing int64
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM contracts c
		JOIN rooms rm ON rm.id = c.room_id
		WHERE c.status = 'active' AND c.start_date <= $5::date AND %s
		AND NOT EXISTS (
			SELECT 1 FROM %s m
			WHERE m.room_id = c.room_id AND m.reading_date >= $6::date AND m.reading_date < $7::date
		)
	`, scopeCondition, table), with(window.Today, window.MonthStart, window.NextMonthStart)...).Scan(&missing)
	return missing, err
}

// readableMenus returns the summary's menu paths the requester's role may
// read. It applies the same rule as the RequirePermission middleware,
// including the active-user check.
func (r *Repository) readableMenus(ctx context.Context, requesterID uuid.UUID) (map[string]bool, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT m.path
		FROM users u
		JOIN role_menu_permissions rmp ON rmp.role_id = u.role_id
		JOIN menus m ON m.id = rmp.menu_id
		JOIN permissions p ON p.id = rmp.permission_id
		WHERE u.id = $1 AND u.is_active = TRUE AND p.name = 'read' AND m.path = ANY($2)
	`, requesterID, []string{
		roomsMenuPath, contractsMenuPath, invoicesMenuPath, repairRequestsMenuPath,
		paymentsMenuPath, expensesMenuPath, metersMenuPath, waterMetersMenuPath,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readable := make(map[string]bool)
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		readable[path] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return readable, nil
}
