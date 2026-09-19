package postgres

import (
	"context"
	"fmt"

	dashboarddomain "apihorpug/internal/features/dashboard/domain"

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
)

// scopeCondition restricts a query on a row aliased rm (rooms) to the
// dormitories the requester manages. $1 is the requester, $2 their role and
// $3 whether the role is exempt from dormitory scoping.
const scopeCondition = `($3 OR rm.dormitory_id IN (
	SELECT dormitory_id FROM user_dormitories WHERE user_id = $1
	UNION
	SELECT dormitory_id FROM role_dormitories WHERE role_id = $2
))`

// GetSummary counts in the database rather than over a fetched page, so the
// figures stay correct however many rows exist.
func (r *Repository) GetSummary(ctx context.Context, requesterID uuid.UUID) (dashboarddomain.Summary, error) {
	var summary dashboarddomain.Summary

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
		`, scopeCondition), requesterID, roleID, full).Scan(&stats.Total, &stats.Available, &stats.Occupied, &stats.Maintenance); err != nil {
			return summary, err
		}
		summary.Rooms = &stats
	}

	if readable[contractsMenuPath] {
		stats := dashboarddomain.ContractStats{}
		if err := r.db.QueryRow(ctx, fmt.Sprintf(`
			SELECT
				COUNT(*),
				COUNT(*) FILTER (WHERE c.status = 'active')
			FROM contracts c
			JOIN rooms rm ON rm.id = c.room_id
			WHERE %s
		`, scopeCondition), requesterID, roleID, full).Scan(&stats.Total, &stats.Active); err != nil {
			return summary, err
		}
		summary.Contracts = &stats
	}

	if readable[invoicesMenuPath] {
		stats := dashboarddomain.InvoiceStats{}
		if err := r.db.QueryRow(ctx, fmt.Sprintf(`
			SELECT
				COUNT(*) FILTER (WHERE i.status IN ('unpaid', 'overdue')),
				COALESCE(SUM(i.total_amount) FILTER (WHERE i.status IN ('unpaid', 'overdue')), 0)::float8
			FROM invoices i
			JOIN contracts c ON c.id = i.contract_id
			JOIN rooms rm ON rm.id = c.room_id
			WHERE %s
		`, scopeCondition), requesterID, roleID, full).Scan(&stats.OutstandingCount, &stats.OutstandingAmount); err != nil {
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
		`, scopeCondition), requesterID, roleID, full).Scan(&stats.Pending, &stats.InProgress); err != nil {
			return summary, err
		}
		summary.Repairs = &stats
	}

	return summary, nil
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
	`, requesterID, []string{roomsMenuPath, contractsMenuPath, invoicesMenuPath, repairRequestsMenuPath})
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
