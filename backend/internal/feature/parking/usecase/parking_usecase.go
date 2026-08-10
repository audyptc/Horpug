package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/parking/domain"
	tenantdomain "apigofiberhorpug/internal/feature/tenant/domain"

	"github.com/google/uuid"
)

type ParkingUseCase struct {
	repo       domain.ParkingSlotRepository
	tenantRepo tenantdomain.TenantRepository
}

func NewParkingUseCase(repo domain.ParkingSlotRepository, tenantRepo tenantdomain.TenantRepository) *ParkingUseCase {
	return &ParkingUseCase{repo: repo, tenantRepo: tenantRepo}
}

func (uc *ParkingUseCase) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.ParkingSlot, int, error) {
	total, err := uc.repo.Count(ctx, dormitoryID)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	list, err := uc.repo.List(ctx, dormitoryID, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return list, total, nil
}

func (uc *ParkingUseCase) GetByID(ctx context.Context, dormitoryID, id string) (*domain.ParkingSlot, error) {
	p, err := uc.repo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return p, nil
}

func (uc *ParkingUseCase) validateTenant(ctx context.Context, dormitoryID string, tenantID *string) error {
	if tenantID == nil || *tenantID == "" {
		return nil
	}
	if _, err := uc.tenantRepo.FindByID(ctx, dormitoryID, *tenantID); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound("tenant not found")
		}
		return apierror.Internal(err)
	}
	return nil
}

func (uc *ParkingUseCase) Create(ctx context.Context, dormitoryID string, req *domain.CreateParkingSlotRequest) (*domain.ParkingSlot, error) {
	if err := uc.validateTenant(ctx, dormitoryID, req.TenantID); err != nil {
		return nil, err
	}
	p := &domain.ParkingSlot{
		ID:           uuid.New().String(),
		DormitoryID:  dormitoryID,
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
	return uc.repo.FindByID(ctx, dormitoryID, p.ID)
}

func (uc *ParkingUseCase) Update(ctx context.Context, dormitoryID, id string, req *domain.UpdateParkingSlotRequest) (*domain.ParkingSlot, error) {
	p, err := uc.repo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	if err := uc.validateTenant(ctx, dormitoryID, req.TenantID); err != nil {
		return nil, err
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
	return uc.repo.FindByID(ctx, dormitoryID, id)
}

func (uc *ParkingUseCase) Delete(ctx context.Context, dormitoryID, id string) error {
	if _, err := uc.repo.FindByID(ctx, dormitoryID, id); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.repo.Delete(ctx, dormitoryID, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
