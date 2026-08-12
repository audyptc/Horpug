package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoomType struct {
	ID            uuid.UUID  `json:"id"`
	DormitoryID   uuid.UUID  `json:"dormitory_id"`
	DormitoryName string     `json:"dormitory_name,omitempty"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Price         float64    `json:"price"`
	IsActive      bool       `json:"is_active"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy     *uuid.UUID `json:"updated_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
