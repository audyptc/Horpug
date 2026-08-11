package domain

import (
	"time"

	"github.com/google/uuid"
)

type Menu struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string    `json:"name" gorm:"size:120;not null"`
	Path        string    `json:"path" gorm:"size:255;uniqueIndex;not null"`
	Description string    `json:"description" gorm:"size:255"`
	IsActive    bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
