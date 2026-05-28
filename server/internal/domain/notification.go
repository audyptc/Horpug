package domain

import (
	"context"
	"time"
)

type NotificationType string

const (
	NotificationTypeOverdueBill       NotificationType = "overdue_bill"
	NotificationTypeExpiringContract  NotificationType = "expiring_contract"
	NotificationTypeOpenMaintenance   NotificationType = "open_maintenance"
)

type NotificationItem struct {
	ID             string           `json:"id"`
	Type           NotificationType `json:"type"`
	TenantName     string           `json:"tenant_name,omitempty"`
	RoomNumber     string           `json:"room_number"`
	TotalAmount    float64          `json:"total_amount,omitempty"`
	DaysOverdue    int              `json:"days_overdue,omitempty"`
	DaysRemaining  int              `json:"days_remaining,omitempty"`
	Title          string           `json:"title,omitempty"`
	Priority       string           `json:"priority,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
}

type NotificationRepository interface {
	ListOverdueBills(ctx context.Context, limit int) ([]*NotificationItem, error)
	ListExpiringContracts(ctx context.Context, withinDays, limit int) ([]*NotificationItem, error)
	ListOpenMaintenance(ctx context.Context, limit int) ([]*NotificationItem, error)
}
