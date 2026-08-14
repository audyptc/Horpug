package domain

import (
	"time"

	"github.com/google/uuid"
)

type ParcelStatus string

const (
	ParcelStatusPending  ParcelStatus = "pending"
	ParcelStatusPickedUp ParcelStatus = "picked_up"
	ParcelStatusReturned ParcelStatus = "returned"
)

func (s ParcelStatus) Valid() bool {
	switch s {
	case ParcelStatusPending, ParcelStatusPickedUp, ParcelStatusReturned:
		return true
	}
	return false
}

// Parcel is a package delivered to the dormitory office on behalf of a
// tenant, tracked from arrival until the tenant picks it up (or it's
// returned to sender), optionally tied to the room the tenant occupies
// (which also determines which dormitory the record belongs to).
type Parcel struct {
	ID             uuid.UUID    `json:"id"`
	TenantID       uuid.UUID    `json:"tenant_id"`
	TenantName     string       `json:"tenant_name,omitempty"`
	RoomID         *uuid.UUID   `json:"room_id,omitempty"`
	RoomNumber     *string      `json:"room_number,omitempty"`
	DormitoryID    *uuid.UUID   `json:"dormitory_id,omitempty"`
	DormitoryName  *string      `json:"dormitory_name,omitempty"`
	Courier        string       `json:"courier"`
	TrackingNumber string       `json:"tracking_number"`
	Status         ParcelStatus `json:"status"`
	ReceivedDate   time.Time    `json:"received_date"`
	Note           string       `json:"note"`
	CreatedBy      *uuid.UUID   `json:"created_by,omitempty"`
	UpdatedBy      *uuid.UUID   `json:"updated_by,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}
