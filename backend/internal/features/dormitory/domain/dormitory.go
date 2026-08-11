package domain

import (
	"time"

	"github.com/google/uuid"
)

type Dormitory struct {
	ID          uuid.UUID          `json:"id"`
	Name        string             `json:"name"`
	Address     string             `json:"address"`
	Phone       string             `json:"phone"`
	Description string             `json:"description"`
	IsActive    bool               `json:"is_active"`
	Managers    []DormitoryManager `json:"managers,omitempty"`
	CreatedBy   *uuid.UUID         `json:"created_by,omitempty"`
	UpdatedBy   *uuid.UUID         `json:"updated_by,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type DormitoryManager struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
