package domain

import (
	"time"

	"github.com/google/uuid"
)

// Dormitory.PromptPayID is the account invoices' PromptPay QR pays into: a
// mobile number, national/tax ID or e-wallet ID, digits only. Empty = no QR.
type Dormitory struct {
	ID          uuid.UUID          `json:"id"`
	Name        string             `json:"name"`
	Address     string             `json:"address"`
	Phone       string             `json:"phone"`
	PromptPayID string             `json:"promptpay_id"`
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
