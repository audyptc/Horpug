package domain

import (
	"time"

	"github.com/google/uuid"
)

type ContractStatus string

const (
	ContractStatusActive     ContractStatus = "active"
	ContractStatusExpired    ContractStatus = "expired"
	ContractStatusTerminated ContractStatus = "terminated"
)

func (s ContractStatus) Valid() bool {
	switch s {
	case ContractStatusActive, ContractStatusExpired, ContractStatusTerminated:
		return true
	}
	return false
}

type Contract struct {
	ID            uuid.UUID      `json:"id"`
	TenantID      uuid.UUID      `json:"tenant_id"`
	TenantName    string         `json:"tenant_name,omitempty"`
	RoomID        uuid.UUID      `json:"room_id"`
	RoomNumber    string         `json:"room_number,omitempty"`
	DormitoryID   uuid.UUID      `json:"dormitory_id,omitempty"`
	DormitoryName string         `json:"dormitory_name,omitempty"`
	StartDate     time.Time      `json:"start_date"`
	EndDate       *time.Time     `json:"end_date,omitempty"`
	RentPrice     float64        `json:"rent_price"`
	Deposit       float64        `json:"deposit"`
	NumOccupants  int            `json:"num_occupants"`
	Status        ContractStatus `json:"status"`
	Note          string         `json:"note"`
	CreatedBy     *uuid.UUID     `json:"created_by,omitempty"`
	UpdatedBy     *uuid.UUID     `json:"updated_by,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
