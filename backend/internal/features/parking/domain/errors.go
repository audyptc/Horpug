package domain

import "errors"

var (
	ErrParkingNotFound     = errors.New("parking registration not found")
	ErrRequiredParkingData = errors.New("tenant_id and license_plate are required")
	ErrInvalidVehicleType  = errors.New("invalid vehicle type")
	ErrTenantNotFound      = errors.New("tenant not found")
)
