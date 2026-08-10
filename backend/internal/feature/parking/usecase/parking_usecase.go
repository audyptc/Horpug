package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/parking/domain"

	"github.com/google/uuid"
)

type ParkingUseCase struct {
	repo domain.ParkingSlotRepository
}

func NewParkingUseCase(repo domain.ParkingSlotRepository) *ParkingUseCase {
	return &ParkingUseCase{repo: repo}
}

func (uc *ParkingUseCase) List(ctx context.Context, limit, offset int) ([]*domain.ParkingSlot, int, error) {
	total, err := uc.repo.Count(ctx)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	list, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return list, total, nil
}

func (uc *ParkingUseCase) GetByID(ctx context.Context, id string) (*domain.ParkingSlot, error) {
	p, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return p, nil
}

func (uc *ParkingUseCase) Create(ctx context.Context, req *domain.CreateParkingSlotRequest) (*domain.ParkingSlot, error) {
	p := &domain.ParkingSlot{
		ID:           uuid.New().String(),
		SlotNumber:   req.SlotNumber,
		VehicleType:  req.VehicleType,
		Status:       req.Status,
		TenantID:     req.TenantID,
		LicensePlate: req.LicensePlate,
		MonthlyFee:   req.MonthlyFee,
		StartDate:    req.StartDate,
		Note:         req.Note,
	}
	if err := uc.repo.Create(ctx, p); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, p.ID)
}

func (uc *ParkingUseCase) Update(ctx context.Context, id string, req *domain.UpdateParkingSlotRequest) (*domain.ParkingSlot, error) {
	p, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	p.SlotNumber = req.SlotNumber
	p.VehicleType = req.VehicleType
	p.Status = req.Status
	p.TenantID = req.TenantID
	p.LicensePlate = req.LicensePlate
	p.MonthlyFee = req.MonthlyFee
	p.StartDate = req.StartDate
	p.Note = req.Note

	if err := uc.repo.Update(ctx, p); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, id)
}

func (uc *ParkingUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
