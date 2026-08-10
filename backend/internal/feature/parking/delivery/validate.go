package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/parking/domain"
)

func validateCreateParkingSlotRequest(req *domain.CreateParkingSlotRequest) error {
	validVehicleTypes := map[domain.VehicleType]bool{
		domain.VehicleTypeCar:        true,
		domain.VehicleTypeMotorcycle: true,
	}
	validStatuses := map[domain.ParkingStatus]bool{
		domain.ParkingStatusAvailable: true,
		domain.ParkingStatusOccupied:  true,
	}
	if req.SlotNumber == "" {
		return apierror.BadRequest("slot_number is required")
	}
	if !validVehicleTypes[req.VehicleType] {
		return apierror.BadRequest("vehicle_type must be one of: car, motorcycle")
	}
	if !validStatuses[req.Status] {
		return apierror.BadRequest("status must be one of: available, occupied")
	}
	if req.MonthlyFee < 0 {
		return apierror.BadRequest("monthly_fee must be >= 0")
	}
	return nil
}

func validateUpdateParkingSlotRequest(req *domain.UpdateParkingSlotRequest) error {
	validVehicleTypes := map[domain.VehicleType]bool{
		domain.VehicleTypeCar:        true,
		domain.VehicleTypeMotorcycle: true,
	}
	validStatuses := map[domain.ParkingStatus]bool{
		domain.ParkingStatusAvailable: true,
		domain.ParkingStatusOccupied:  true,
	}
	if req.SlotNumber == "" {
		return apierror.BadRequest("slot_number is required")
	}
	if !validVehicleTypes[req.VehicleType] {
		return apierror.BadRequest("vehicle_type must be one of: car, motorcycle")
	}
	if !validStatuses[req.Status] {
		return apierror.BadRequest("status must be one of: available, occupied")
	}
	if req.MonthlyFee < 0 {
		return apierror.BadRequest("monthly_fee must be >= 0")
	}
	return nil
}
