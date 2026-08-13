package domain

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID               uuid.UUID  `json:"id"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	Phone            string     `json:"phone"`
	IDCard           string     `json:"id_card"`
	Email            string     `json:"email"`
	EmergencyContact string     `json:"emergency_contact"`
	Note             string     `json:"note"`
	IsActive         bool       `json:"is_active"`
	CreatedBy        *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy        *uuid.UUID `json:"updated_by,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
