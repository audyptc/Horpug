package domain

import (
	"context"
	"time"
)

type ContractStatus string

const (
	ContractStatusActive     ContractStatus = "active"
	ContractStatusExpired    ContractStatus = "expired"
	ContractStatusTerminated ContractStatus = "terminated"
)

type Contract struct {
	ID           string         `json:"id"`
	TenantID     string         `json:"tenant_id"`
	RoomID       string         `json:"room_id"`
	StartDate    time.Time      `json:"start_date"`
	EndDate      *time.Time     `json:"end_date"`
	RentPrice    float64        `json:"rent_price"`
	Deposit      float64        `json:"deposit"`
	NumOccupants int            `json:"num_occupants"`
	Status       ContractStatus `json:"status"`
	Note         string         `json:"note"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	CreatedBy    string         `json:"created_by"`
	UpdatedBy    string         `json:"updated_by"`
}

type ContractDetail struct {
	Contract
	TenantFirstName string `json:"tenant_first_name"`
	TenantLastName  string `json:"tenant_last_name"`
	RoomNumber      string `json:"room_number"`
	UpdatedByName   string `json:"updated_by_name"`
}

type CreateContractRequest struct {
	TenantID     string     `json:"tenant_id"`
	RoomID       string     `json:"room_id"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	RentPrice    float64    `json:"rent_price"`
	Deposit      float64    `json:"deposit"`
	NumOccupants int        `json:"num_occupants"`
	Note         string     `json:"note"`
}

type UpdateContractRequest struct {
	EndDate      *time.Time     `json:"end_date"`
	RentPrice    float64        `json:"rent_price"`
	Deposit      float64        `json:"deposit"`
	NumOccupants *int           `json:"num_occupants"`
	Status       ContractStatus `json:"status"`
	Note         string         `json:"note"`
}

type ContractRepository interface {
	// FindByID and Update/Delete are scoped by dormitoryID via a join to rooms
	// (contracts don't carry their own dormitory_id column).
	FindByID(ctx context.Context, dormitoryID, id string) (*Contract, error)
	FindDetailByID(ctx context.Context, dormitoryID, id string) (*ContractDetail, error)
	List(ctx context.Context, dormitoryID string, limit, offset int) ([]*ContractDetail, error)
	Count(ctx context.Context, dormitoryID string) (int, error)
	Create(ctx context.Context, c *Contract) error
	Update(ctx context.Context, dormitoryID string, c *Contract) error
	Delete(ctx context.Context, dormitoryID, id string) error
	HasActiveContractForRoom(ctx context.Context, roomID string) (bool, error)
	HasActiveContractForTenant(ctx context.Context, tenantID string) (bool, error)

	// FindByIDAny looks up a contract without dormitory scoping, for features
	// that are not yet dormitory-scoped themselves (bill). Prefer FindByID
	// wherever a dormitory context is available.
	FindByIDAny(ctx context.Context, id string) (*Contract, error)
}
