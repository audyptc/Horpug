package domain

import (
	"time"

	"github.com/google/uuid"
)

type ActivityLog struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	UserID      *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid"`
	Username    string     `json:"username,omitempty" gorm:"-"`
	Action      string     `json:"action" gorm:"size:50;not null"`
	EntityType  string     `json:"entity_type" gorm:"size:80;not null"`
	EntityID    *uuid.UUID `json:"entity_id,omitempty" gorm:"type:uuid"`
	Description string     `json:"description" gorm:"size:255"`
	IPAddress   string     `json:"ip_address" gorm:"size:45"`
	CreatedAt   time.Time  `json:"created_at"`
}
