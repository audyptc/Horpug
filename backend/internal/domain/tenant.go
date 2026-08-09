package domain

import (
	"context"
	"time"
)

type Tenant struct {
	ID               string    `json:"id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Phone            string    `json:"phone"`
	IDCard           string    `json:"id_card"`
	Email            string    `json:"email"`
	EmergencyContact string    `json:"emergency_contact"`
	Note             string    `json:"note"`
	ActiveRoomNumbers []string  `json:"active_room_numbers"`
	CreatedBy     string    `json:"created_by"`
	UpdatedBy     string    `json:"updated_by"`
	UpdatedByName string    `json:"updated_by_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateTenantRequest struct {
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Phone            string `json:"phone"`
	IDCard           string `json:"id_card"`
	Email            string `json:"email"`
	EmergencyContact string `json:"emergency_contact"`
	Note             string `json:"note"`
}

type UpdateTenantRequest struct {
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Phone            string `json:"phone"`
	IDCard           string `json:"id_card"`
	Email            string `json:"email"`
	EmergencyContact string `json:"emergency_contact"`
	Note             string `json:"note"`
}

type TenantRepository interface {
	FindByID(ctx context.Context, id string) (*Tenant, error)
	List(ctx context.Context, limit, offset int) ([]*Tenant, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, tenant *Tenant) error
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id string) error
}
