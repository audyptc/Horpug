package domain

import (
	"context"
	"time"
)

type NotificationType string

const (
	NotificationTypeOverdueBill NotificationType = "overdue_bill"
)

type NotificationItem struct {
	ID          string           `json:"id"`
	Type        NotificationType `json:"type"`
	TenantName  string           `json:"tenant_name"`
	RoomNumber  string           `json:"room_number"`
	TotalAmount float64          `json:"total_amount"`
	DaysOverdue int              `json:"days_overdue"`
	BillID      string           `json:"bill_id"`
	CreatedAt   time.Time        `json:"created_at"`
}

type NotificationRepository interface {
	ListOverdueBills(ctx context.Context, limit int) ([]*NotificationItem, error)
}
