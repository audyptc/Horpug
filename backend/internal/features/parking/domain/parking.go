package domain

import (
	"context"
	"time"
)

type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
)

type ParkingStatus string

const (
	ParkingStatusAvailable ParkingStatus = "available"
	ParkingStatusOccupied  ParkingStatus = "occupied"
)

type ParkingSlot struct {
	ID           string        `json:"id"`
	DormitoryID  string        `json:"dormitory_id"`
	SlotNumber   string        `json:"slot_number"`
	VehicleType  VehicleType   `json:"vehicle_type"`
	Status       ParkingStatus `json:"status"`
	TenantID     *string       `json:"tenant_id"`
	LicensePlate string        `json:"license_plate"`
	MonthlyFee   float64       `json:"monthly_fee"`
	StartDate    *time.Time    `json:"start_date"`
	Note         string        `json:"note"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	// joined
	TenantFirstName string `json:"tenant_first_name"`
	TenantLastName  string `json:"tenant_last_name"`
}

type CreateParkingSlotRequest struct {
	SlotNumber   string        `json:"slot_number"`
	VehicleType  VehicleType   `json:"vehicle_type"`
	Status       ParkingStatus `json:"status"`
	TenantID     *string       `json:"tenant_id"`
	LicensePlate string        `json:"license_plate"`
	MonthlyFee   float64       `json:"monthly_fee"`
	StartDate    *time.Time    `json:"start_date"`
	Note         string        `json:"note"`
}

type UpdateParkingSlotRequest struct {
	SlotNumber   string        `json:"slot_number"`
	VehicleType  VehicleType   `json:"vehicle_type"`
	Status       ParkingStatus `json:"status"`
	TenantID     *string       `json:"tenant_id"`
	LicensePlate string        `json:"license_plate"`
	MonthlyFee   float64       `json:"monthly_fee"`
	StartDate    *time.Time    `json:"start_date"`
	Note         string        `json:"note"`
}

type ParkingSlotRepository interface {
	FindByID(ctx context.Context, dormitoryID, id string) (*ParkingSlot, error)
	List(ctx context.Context, dormitoryID string, limit, offset int) ([]*ParkingSlot, error)
	Count(ctx context.Context, dormitoryID string) (int, error)
	Create(ctx context.Context, p *ParkingSlot) error
	Update(ctx context.Context, p *ParkingSlot) error
	Delete(ctx context.Context, dormitoryID, id string) error
}
