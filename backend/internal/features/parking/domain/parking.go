package domain

import (
	"time"

	"github.com/google/uuid"
)

type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
	VehicleTypeOther      VehicleType = "other"
)

func (v VehicleType) Valid() bool {
	switch v {
	case VehicleTypeCar, VehicleTypeMotorcycle, VehicleTypeOther:
		return true
	}
	return false
}

// Parking is a tenant's registered vehicle and its assigned parking spot.
type Parking struct {
	ID           uuid.UUID   `json:"id"`
	TenantID     uuid.UUID   `json:"tenant_id"`
	TenantName   string      `json:"tenant_name,omitempty"`
	VehicleType  VehicleType `json:"vehicle_type"`
	LicensePlate string      `json:"license_plate"`
	ParkingSpot  string      `json:"parking_spot"`
	CreatedBy    *uuid.UUID  `json:"created_by,omitempty"`
	UpdatedBy    *uuid.UUID  `json:"updated_by,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
