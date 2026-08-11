package domain

import (
	"context"
	"time"
)

type MaintenanceStatus string
type MaintenancePriority string

const (
	MaintenanceStatusOpen       MaintenanceStatus = "open"
	MaintenanceStatusInProgress MaintenanceStatus = "in_progress"
	MaintenanceStatusDone       MaintenanceStatus = "done"
	MaintenanceStatusCancelled  MaintenanceStatus = "cancelled"
)

const (
	MaintenancePriorityLow    MaintenancePriority = "low"
	MaintenancePriorityNormal MaintenancePriority = "normal"
	MaintenancePriorityHigh   MaintenancePriority = "high"
	MaintenancePriorityUrgent MaintenancePriority = "urgent"
)

type MaintenanceRequest struct {
	ID           string              `json:"id"`
	RoomID       string              `json:"room_id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	Status       MaintenanceStatus   `json:"status"`
	Priority     MaintenancePriority `json:"priority"`
	ReportedDate time.Time           `json:"reported_date"`
	ResolvedDate *time.Time          `json:"resolved_date"`
	Note         string              `json:"note"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

type MaintenanceRequestDetail struct {
	MaintenanceRequest
	RoomNumber string `json:"room_number"`
}

type CreateMaintenanceRequestRequest struct {
	RoomID       string              `json:"room_id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	Status       MaintenanceStatus   `json:"status"`
	Priority     MaintenancePriority `json:"priority"`
	ReportedDate time.Time           `json:"reported_date"`
	ResolvedDate *time.Time          `json:"resolved_date"`
	Note         string              `json:"note"`
}

type UpdateMaintenanceRequestRequest struct {
	RoomID       string              `json:"room_id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	Status       MaintenanceStatus   `json:"status"`
	Priority     MaintenancePriority `json:"priority"`
	ReportedDate time.Time           `json:"reported_date"`
	ResolvedDate *time.Time          `json:"resolved_date"`
	Note         string              `json:"note"`
}

type MaintenanceRequestRepository interface {
	FindByID(ctx context.Context, dormitoryID, id string) (*MaintenanceRequest, error)
	FindDetailByID(ctx context.Context, dormitoryID, id string) (*MaintenanceRequestDetail, error)
	List(ctx context.Context, dormitoryID string, limit, offset int) ([]*MaintenanceRequestDetail, error)
	Count(ctx context.Context, dormitoryID string) (int, error)
	Create(ctx context.Context, m *MaintenanceRequest) error
	Update(ctx context.Context, dormitoryID string, m *MaintenanceRequest) error
	Delete(ctx context.Context, dormitoryID, id string) error
}
