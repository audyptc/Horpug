package domain

import (
	"context"
	"encoding/json"
	"time"
)

type ActivityAction string

const (
	ActivityCreate ActivityAction = "create"
	ActivityUpdate ActivityAction = "update"
	ActivityDelete ActivityAction = "delete"
)

type ActivityLog struct {
	ID          string          `json:"id"`
	ActorID     *string         `json:"actor_id"`
	ActorName   string          `json:"actor_name,omitempty"`
	DormitoryID *string         `json:"dormitory_id,omitempty"`
	Action      ActivityAction  `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    string          `json:"entity_id"`
	NewValue    json.RawMessage `json:"new_value,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type ActivityLogFilter struct {
	DormitoryID string
	EntityType  string
	ActorID     string
	From        *time.Time
	To          *time.Time
}

type ActivityLogRepository interface {
	Create(ctx context.Context, log *ActivityLog) error
	List(ctx context.Context, filter ActivityLogFilter, limit, offset int) ([]*ActivityLog, error)
	Count(ctx context.Context, filter ActivityLogFilter) (int, error)
}
