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
	ID        string         `json:"id"`
	TenantID  string         `json:"tenant_id"`
	RoomID    string         `json:"room_id"`
	StartDate time.Time      `json:"start_date"`
	EndDate   *time.Time     `json:"end_date"`
	RentPrice float64        `json:"rent_price"`
	Deposit   float64        `json:"deposit"`
	Status    ContractStatus `json:"status"`
	Note      string         `json:"note"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type ContractDetail struct {
	Contract
	TenantFirstName string `json:"tenant_first_name"`
	TenantLastName  string `json:"tenant_last_name"`
	RoomNumber      string `json:"room_number"`
}

type CreateContractRequest struct {
	TenantID  string     `json:"tenant_id"`
	RoomID    string     `json:"room_id"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	RentPrice float64    `json:"rent_price"`
	Deposit   float64    `json:"deposit"`
	Note      string     `json:"note"`
}

type UpdateContractRequest struct {
	EndDate   *time.Time     `json:"end_date"`
	RentPrice float64        `json:"rent_price"`
	Deposit   float64        `json:"deposit"`
	Status    ContractStatus `json:"status"`
	Note      string         `json:"note"`
}

type ContractRepository interface {
	FindByID(ctx context.Context, id string) (*Contract, error)
	FindDetailByID(ctx context.Context, id string) (*ContractDetail, error)
	List(ctx context.Context, limit, offset int) ([]*ContractDetail, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, c *Contract) error
	Update(ctx context.Context, c *Contract) error
	Delete(ctx context.Context, id string) error
	HasActiveContractForRoom(ctx context.Context, roomID string) (bool, error)
	HasActiveContractForTenant(ctx context.Context, tenantID string) (bool, error)
}
