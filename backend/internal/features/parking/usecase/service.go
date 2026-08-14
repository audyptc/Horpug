package usecase

import (
	"context"
	"strings"

	parkingdomain "apihorpug/internal/features/parking/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	TenantID    *uuid.UUID
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
	VehicleType *parkingdomain.VehicleType
}

type CreateInput struct {
	TenantID     uuid.UUID
	RoomID       *uuid.UUID
	VehicleType  parkingdomain.VehicleType
	LicensePlate string
	ParkingSpot  string
	CreatedBy    *uuid.UUID
}

type UpdateInput struct {
	RoomID       *uuid.UUID
	VehicleType  *parkingdomain.VehicleType
	LicensePlate *string
	ParkingSpot  *string
	UpdatedBy    *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]parkingdomain.Parking, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (parkingdomain.Parking, error)
	Create(ctx context.Context, input CreateInput) (parkingdomain.Parking, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (parkingdomain.Parking, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]parkingdomain.Parking, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	parkings, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return parkings, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (parkingdomain.Parking, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (parkingdomain.Parking, error) {
	input.LicensePlate = strings.TrimSpace(input.LicensePlate)
	input.ParkingSpot = strings.TrimSpace(input.ParkingSpot)
	if input.VehicleType == "" {
		input.VehicleType = parkingdomain.VehicleTypeMotorcycle
	}
	if input.RoomID != nil && *input.RoomID == uuid.Nil {
		input.RoomID = nil
	}

	if input.TenantID == uuid.Nil || input.LicensePlate == "" {
		return parkingdomain.Parking{}, parkingdomain.ErrRequiredParkingData
	}
	if !input.VehicleType.Valid() {
		return parkingdomain.Parking{}, parkingdomain.ErrInvalidVehicleType
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (parkingdomain.Parking, error) {
	if input.VehicleType != nil && !input.VehicleType.Valid() {
		return parkingdomain.Parking{}, parkingdomain.ErrInvalidVehicleType
	}
	if input.LicensePlate != nil {
		licensePlate := strings.TrimSpace(*input.LicensePlate)
		if licensePlate == "" {
			return parkingdomain.Parking{}, parkingdomain.ErrRequiredParkingData
		}
		input.LicensePlate = &licensePlate
	}
	if input.ParkingSpot != nil {
		parkingSpot := strings.TrimSpace(*input.ParkingSpot)
		input.ParkingSpot = &parkingSpot
	}

	return s.repo.Update(ctx, id, requesterID, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
