package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"

	"github.com/google/uuid"
)

type MeterReadingUseCase struct {
	meterRepo domain.MeterReadingRepository
	roomRepo  domain.RoomRepository
}

func NewMeterReadingUseCase(meterRepo domain.MeterReadingRepository, roomRepo domain.RoomRepository) *MeterReadingUseCase {
	return &MeterReadingUseCase{meterRepo: meterRepo, roomRepo: roomRepo}
}

func (uc *MeterReadingUseCase) List(ctx context.Context, limit, offset int) ([]*domain.MeterReadingDetail, int, error) {
	total, err := uc.meterRepo.Count(ctx)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	list, err := uc.meterRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return list, total, nil
}

func (uc *MeterReadingUseCase) GetByID(ctx context.Context, id string) (*domain.MeterReadingDetail, error) {
	d, err := uc.meterRepo.FindDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return d, nil
}

func (uc *MeterReadingUseCase) Create(ctx context.Context, req *domain.CreateMeterReadingRequest) (*domain.MeterReadingDetail, error) {
	if _, err := uc.roomRepo.FindByID(ctx, req.RoomID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound("room not found")
		}
		return nil, apierror.Internal(err)
	}

	m := &domain.MeterReading{
		ID:              uuid.New().String(),
		RoomID:          req.RoomID,
		MeterType:       req.MeterType,
		ReadingDate:     req.ReadingDate,
		PreviousReading: req.PreviousReading,
		CurrentReading:  req.CurrentReading,
		UnitPrice:       req.UnitPrice,
		Note:            req.Note,
	}
	if err := uc.meterRepo.Create(ctx, m); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.meterRepo.FindDetailByID(ctx, m.ID)
}

func (uc *MeterReadingUseCase) Update(ctx context.Context, id string, req *domain.UpdateMeterReadingRequest) (*domain.MeterReadingDetail, error) {
	m, err := uc.meterRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	if !req.ReadingDate.IsZero() {
		m.ReadingDate = req.ReadingDate
	}
	m.PreviousReading = req.PreviousReading
	m.CurrentReading = req.CurrentReading
	if req.UnitPrice > 0 {
		m.UnitPrice = req.UnitPrice
	}
	m.Note = req.Note

	if err := uc.meterRepo.Update(ctx, m); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.meterRepo.FindDetailByID(ctx, id)
}

func (uc *MeterReadingUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.meterRepo.FindByID(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.meterRepo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
