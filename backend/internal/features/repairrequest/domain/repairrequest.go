package domain

import (
	"time"

	"github.com/google/uuid"
)

type RepairCategory string

const (
	RepairCategoryElectrical RepairCategory = "electrical"
	RepairCategoryPlumbing   RepairCategory = "plumbing"
	RepairCategoryFurniture  RepairCategory = "furniture"
	RepairCategoryAircon     RepairCategory = "aircon"
	RepairCategoryOther      RepairCategory = "other"
)

func (c RepairCategory) Valid() bool {
	switch c {
	case RepairCategoryElectrical, RepairCategoryPlumbing, RepairCategoryFurniture, RepairCategoryAircon, RepairCategoryOther:
		return true
	}
	return false
}

type RepairStatus string

const (
	RepairStatusPending    RepairStatus = "pending"
	RepairStatusInProgress RepairStatus = "in_progress"
	RepairStatusCompleted  RepairStatus = "completed"
	RepairStatusCancelled  RepairStatus = "cancelled"
)

func (s RepairStatus) Valid() bool {
	switch s {
	case RepairStatusPending, RepairStatusInProgress, RepairStatusCompleted, RepairStatusCancelled:
		return true
	}
	return false
}

// RepairRequest is a tenant- or staff-reported maintenance issue tied to a
// room (e.g. broken outlet, leaking pipe).
type RepairRequest struct {
	ID            uuid.UUID      `json:"id"`
	RoomID        uuid.UUID      `json:"room_id"`
	RoomNumber    string         `json:"room_number,omitempty"`
	DormitoryID   uuid.UUID      `json:"dormitory_id,omitempty"`
	DormitoryName string         `json:"dormitory_name,omitempty"`
	TenantID      *uuid.UUID     `json:"tenant_id,omitempty"`
	TenantName    *string        `json:"tenant_name,omitempty"`
	Category      RepairCategory `json:"category"`
	Description   string         `json:"description"`
	Status        RepairStatus   `json:"status"`
	ReportedDate  time.Time      `json:"reported_date"`
	CreatedBy     *uuid.UUID     `json:"created_by,omitempty"`
	UpdatedBy     *uuid.UUID     `json:"updated_by,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
