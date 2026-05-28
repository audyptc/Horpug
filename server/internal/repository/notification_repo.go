package repository

import (
	"context"
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
			BillID:      id,
			CreatedAt:   createdAt,
		})
	}
	if list == nil {
		list = []*domain.NotificationItem{}
	}
	return list, rows.Err()
}
