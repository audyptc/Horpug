package repository

import (
	"context"
	"fmt"
	"time"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"
)

type NotificationRepo struct {
	db *database.DB
}

func NewNotificationRepo(db *database.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) ListOverdueBills(ctx context.Context, limit int) ([]*domain.NotificationItem, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			b.id, b.total_amount, b.due_date, b.created_at,
			t.first_name, t.last_name, rm.room_number
		FROM bills b
		JOIN contracts c  ON c.id = b.contract_id
		JOIN tenants t    ON t.id = c.tenant_id
		JOIN rooms rm     ON rm.id = c.room_id
		WHERE b.status = 'overdue'
		   OR (b.status = 'unpaid' AND b.due_date IS NOT NULL AND b.due_date < NOW())
		ORDER BY b.due_date ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	var list []*domain.NotificationItem
	for rows.Next() {
		var (
			id, firstName, lastName, roomNumber string
			totalAmount                         float64
			dueDate                             *time.Time
			createdAt                           time.Time
		)
		if err := rows.Scan(&id, &totalAmount, &dueDate, &createdAt, &firstName, &lastName, &roomNumber); err != nil {
			return nil, err
		}

		daysOverdue := 0
		if dueDate != nil && dueDate.Before(now) {
			daysOverdue = int(now.Sub(*dueDate).Hours() / 24)
		}

		list = append(list, &domain.NotificationItem{
			ID:          id,
			Type:        domain.NotificationTypeOverdueBill,
			TenantName:  firstName + " " + lastName,
			RoomNumber:  roomNumber,
			TotalAmount: totalAmount,
			DaysOverdue: daysOverdue,
			CreatedAt:   createdAt,
		})
	}
	if list == nil {
		list = []*domain.NotificationItem{}
	}
	return list, rows.Err()
}

func (r *NotificationRepo) ListExpiringContracts(ctx context.Context, withinDays, limit int) ([]*domain.NotificationItem, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			c.id, c.end_date, c.created_at,
			t.first_name, t.last_name, rm.room_number,
			GREATEST(0, EXTRACT(DAY FROM c.end_date - NOW())::int) AS days_remaining
		FROM contracts c
		JOIN tenants t  ON t.id = c.tenant_id
		JOIN rooms rm   ON rm.id = c.room_id
		WHERE c.status = 'active'
		  AND c.end_date IS NOT NULL
		  AND c.end_date BETWEEN NOW() AND NOW() + ($1 || ' days')::interval
		ORDER BY c.end_date ASC
		LIMIT $2`, withinDays, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.NotificationItem
	for rows.Next() {
		var (
			id, firstName, lastName, roomNumber string
			endDate                             *time.Time
			createdAt                           time.Time
			daysRemaining                       int
		)
		if err := rows.Scan(&id, &endDate, &createdAt, &firstName, &lastName, &roomNumber, &daysRemaining); err != nil {
			return nil, err
		}

		list = append(list, &domain.NotificationItem{
			ID:            id,
			Type:          domain.NotificationTypeExpiringContract,
			TenantName:    firstName + " " + lastName,
			RoomNumber:    roomNumber,
			DaysRemaining: daysRemaining,
			CreatedAt:     createdAt,
		})
	}
	if list == nil {
		list = []*domain.NotificationItem{}
	}
	return list, rows.Err()
}

func (r *NotificationRepo) ListOpenMaintenance(ctx context.Context, limit int) ([]*domain.NotificationItem, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			m.id, m.title, m.priority, m.created_at,
			rm.room_number
		FROM maintenance_requests m
		JOIN rooms rm ON rm.id = m.room_id
		WHERE m.status IN ('open', 'in_progress')
		ORDER BY
			CASE m.priority
				WHEN 'urgent' THEN 1
				WHEN 'high'   THEN 2
				WHEN 'normal' THEN 3
				WHEN 'low'    THEN 4
			END,
			m.reported_date ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.NotificationItem
	for rows.Next() {
		var (
			id, title, priority, roomNumber string
			createdAt                       time.Time
		)
		if err := rows.Scan(&id, &title, &priority, &createdAt, &roomNumber); err != nil {
			return nil, err
		}

		list = append(list, &domain.NotificationItem{
			ID:         id,
			Type:       domain.NotificationTypeOpenMaintenance,
			RoomNumber: roomNumber,
			Title:      fmt.Sprintf("[%s] %s", priority, title),
			Priority:   priority,
			CreatedAt:  createdAt,
		})
	}
	if list == nil {
		list = []*domain.NotificationItem{}
	}
	return list, rows.Err()
}
