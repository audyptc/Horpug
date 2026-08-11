package usecase

import (
	"context"
	"errors"

	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/parcel/domain"
	"apigofiberhorpug/internal/shared/http/apierror"

	"github.com/google/uuid"
)

type ParcelUseCase struct {
	repo domain.ParcelRepository
}

func NewParcelUseCase(repo domain.ParcelRepository) *ParcelUseCase {
	return &ParcelUseCase{repo: repo}
}

func (uc *ParcelUseCase) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.Parcel, int, error) {
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

func (uc *ParcelUseCase) GetByID(ctx context.Context, dormitoryID, id string) (*domain.Parcel, error) {
	p, err := uc.repo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return p, nil
}

func (uc *ParcelUseCase) Create(ctx context.Context, dormitoryID string, req *domain.CreateParcelRequest) (*domain.Parcel, error) {
	p := &domain.Parcel{
		ID:             uuid.New().String(),
		DormitoryID:    dormitoryID,
		TrackingNumber: req.TrackingNumber,
		RecipientName:  req.RecipientName,
		RoomNumber:     req.RoomNumber,
		Status:         req.Status,
		ReceivedDate:   req.ReceivedDate,
		PickedUpDate:   req.PickedUpDate,
		Note:           req.Note,
	}
	if err := uc.repo.Create(ctx, p); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, dormitoryID, p.ID)
}

func (uc *ParcelUseCase) Update(ctx context.Context, dormitoryID, id string, req *domain.UpdateParcelRequest) (*domain.Parcel, error) {
	p, err := uc.repo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	p.TrackingNumber = req.TrackingNumber
	p.RecipientName = req.RecipientName
	p.RoomNumber = req.RoomNumber
	p.Status = req.Status
	p.ReceivedDate = req.ReceivedDate
	p.PickedUpDate = req.PickedUpDate
	p.Note = req.Note

	if err := uc.repo.Update(ctx, p); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, dormitoryID, id)
}

func (uc *ParcelUseCase) Delete(ctx context.Context, dormitoryID, id string) error {
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
