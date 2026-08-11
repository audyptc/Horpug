package usecase

import (
	"context"
	"errors"
	"time"

	"apigofiberhorpug/internal/features/electricmeter/domain"
	roomdomain "apigofiberhorpug/internal/features/room/domain"
	coredomain "apigofiberhorpug/internal/shared/domain"
	"apigofiberhorpug/internal/shared/http/apierror"

	"github.com/google/uuid"
)

type ElectricMeterUseCase struct {
	repo     domain.ElectricMeterRepository
	roomRepo roomdomain.RoomRepository
}

func NewElectricMeterUseCase(repo domain.ElectricMeterRepository, roomRepo roomdomain.RoomRepository) *ElectricMeterUseCase {
	return &ElectricMeterUseCase{repo: repo, roomRepo: roomRepo}
}

func (uc *ElectricMeterUseCase) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.ElectricMeterDetail, int, error) {
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

func (uc *ElectricMeterUseCase) GetByID(ctx context.Context, dormitoryID, id string) (*domain.ElectricMeterDetail, error) {
	d, err := uc.repo.FindDetailByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return d, nil
}

func (uc *ElectricMeterUseCase) GetLatestByRoomID(ctx context.Context, dormitoryID, roomID string, billingMonth *time.Time) (*domain.ElectricMeterDetail, error) {
	d, err := uc.repo.FindLatestByRoomID(ctx, dormitoryID, roomID, billingMonth)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return d, nil
}

func (uc *ElectricMeterUseCase) Create(ctx context.Context, dormitoryID string, req *domain.CreateElectricMeterRequest, actorID string) (*domain.ElectricMeterDetail, error) {
	if _, err := uc.roomRepo.FindByID(ctx, dormitoryID, req.RoomID); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound("room not found")
		}
		return nil, apierror.Internal(err)
	}

	m := &domain.ElectricMeter{
		ID:              uuid.New().String(),
		RoomID:          req.RoomID,
		BillingType:     req.BillingType,
		BillingMonth:    req.BillingMonth,
		ReadingDate:     req.ReadingDate,
		PreviousReading: req.PreviousReading,
		CurrentReading:  req.CurrentReading,
		UnitPrice:       req.UnitPrice,
		FlatAmount:      req.FlatAmount,
		Note:            req.Note,
		CreatedBy:       actorID,
		UpdatedBy:       actorID,
	}
	if err := uc.repo.Create(ctx, m); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindDetailByID(ctx, dormitoryID, m.ID)
}

func (uc *ElectricMeterUseCase) Update(ctx context.Context, dormitoryID, id string, req *domain.UpdateElectricMeterRequest, actorID string) (*domain.ElectricMeterDetail, error) {
	m, err := uc.repo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	if !req.ReadingDate.IsZero() {
		m.ReadingDate = req.ReadingDate
	}
	m.BillingType = req.BillingType
	m.BillingMonth = req.BillingMonth
	m.PreviousReading = req.PreviousReading
	m.CurrentReading = req.CurrentReading
	m.UnitPrice = req.UnitPrice
	m.FlatAmount = req.FlatAmount
	m.Note = req.Note
	m.UpdatedBy = actorID

	if err := uc.repo.Update(ctx, dormitoryID, m); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindDetailByID(ctx, dormitoryID, id)
}

func (uc *ElectricMeterUseCase) Delete(ctx context.Context, dormitoryID, id string) error {
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
