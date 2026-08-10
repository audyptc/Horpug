package usecase

import (
	"context"
	"errors"
	"time"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	roomdomain "apigofiberhorpug/internal/feature/room/domain"
	"apigofiberhorpug/internal/feature/watermeter/domain"

	"github.com/google/uuid"
)

type WaterMeterUseCase struct {
	repo     domain.WaterMeterRepository
	roomRepo roomdomain.RoomRepository
}

func NewWaterMeterUseCase(repo domain.WaterMeterRepository, roomRepo roomdomain.RoomRepository) *WaterMeterUseCase {
	return &WaterMeterUseCase{repo: repo, roomRepo: roomRepo}
}

func (uc *WaterMeterUseCase) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.WaterMeterDetail, int, error) {
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

func (uc *WaterMeterUseCase) GetByID(ctx context.Context, dormitoryID, id string) (*domain.WaterMeterDetail, error) {
	d, err := uc.repo.FindDetailByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return d, nil
}

func (uc *WaterMeterUseCase) GetLatestByRoomID(ctx context.Context, dormitoryID, roomID string, billingMonth *time.Time) (*domain.WaterMeterDetail, error) {
	d, err := uc.repo.FindLatestByRoomID(ctx, dormitoryID, roomID, billingMonth)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return d, nil
}

func (uc *WaterMeterUseCase) Create(ctx context.Context, dormitoryID string, req *domain.CreateWaterMeterRequest, actorID string) (*domain.WaterMeterDetail, error) {
	if _, err := uc.roomRepo.FindByID(ctx, dormitoryID, req.RoomID); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound("room not found")
		}
		return nil, apierror.Internal(err)
	}

	m := &domain.WaterMeter{
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

func (uc *WaterMeterUseCase) Update(ctx context.Context, dormitoryID, id string, req *domain.UpdateWaterMeterRequest, actorID string) (*domain.WaterMeterDetail, error) {
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

func (uc *WaterMeterUseCase) Delete(ctx context.Context, dormitoryID, id string) error {
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
