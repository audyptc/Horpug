package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"

	"github.com/google/uuid"
)

type WaterMeterUseCase struct {
	repo     domain.WaterMeterRepository
	roomRepo domain.RoomRepository
}

func NewWaterMeterUseCase(repo domain.WaterMeterRepository, roomRepo domain.RoomRepository) *WaterMeterUseCase {
	return &WaterMeterUseCase{repo: repo, roomRepo: roomRepo}
}

func (uc *WaterMeterUseCase) List(ctx context.Context, limit, offset int) ([]*domain.WaterMeterDetail, int, error) {
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

func (uc *WaterMeterUseCase) GetByID(ctx context.Context, id string) (*domain.WaterMeterDetail, error) {
	d, err := uc.repo.FindDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return d, nil
}

func (uc *WaterMeterUseCase) Create(ctx context.Context, req *domain.CreateWaterMeterRequest, actorID string) (*domain.WaterMeterDetail, error) {
	if _, err := uc.roomRepo.FindByID(ctx, req.RoomID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound("room not found")
		}
		return nil, apierror.Internal(err)
	}

	m := &domain.WaterMeter{
		ID:              uuid.New().String(),
		RoomID:          req.RoomID,
		BillingType:     req.BillingType,
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
	return uc.repo.FindDetailByID(ctx, m.ID)
}

func (uc *WaterMeterUseCase) Update(ctx context.Context, id string, req *domain.UpdateWaterMeterRequest, actorID string) (*domain.WaterMeterDetail, error) {
	m, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	if !req.ReadingDate.IsZero() {
		m.ReadingDate = req.ReadingDate
	}
	m.BillingType = req.BillingType
	m.PreviousReading = req.PreviousReading
	m.CurrentReading = req.CurrentReading
	m.UnitPrice = req.UnitPrice
	m.FlatAmount = req.FlatAmount
	m.Note = req.Note
	m.UpdatedBy = actorID

	if err := uc.repo.Update(ctx, m); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindDetailByID(ctx, id)
}

func (uc *WaterMeterUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
