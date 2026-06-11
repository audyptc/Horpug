package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"

	"github.com/google/uuid"
)

type RoomUseCase struct {
	roomRepo domain.RoomRepository
}

func NewRoomUseCase(roomRepo domain.RoomRepository) *RoomUseCase {
	return &RoomUseCase{roomRepo: roomRepo}
}

func (uc *RoomUseCase) List(ctx context.Context, limit, offset int) ([]*domain.Room, int, error) {
	total, err := uc.roomRepo.Count(ctx)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	rooms, err := uc.roomRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return rooms, total, nil
}

func (uc *RoomUseCase) GetByID(ctx context.Context, id string) (*domain.Room, error) {
	room, err := uc.roomRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return room, nil
}

func (uc *RoomUseCase) Create(ctx context.Context, req *domain.CreateRoomRequest) (*domain.Room, error) {
	room := &domain.Room{
		ID:          uuid.New().String(),
		RoomNumber:  req.RoomNumber,
		Floor:       req.Floor,
		Type:        req.Type,
		Status:      req.Status,
		RentPrice:   req.RentPrice,
		Description: req.Description,
	}
	if room.Type == "" {
		room.Type = "standard"
	}
	if room.Status == "" {
		room.Status = "available"
	}
	if err := uc.roomRepo.Create(ctx, room); err != nil {
		if errors.Is(err, domain.ErrDuplicate) {
			return nil, apierror.Conflict("ห้องหมายเลขนี้มีอยู่แล้ว")
		}
		return nil, apierror.Internal(err)
	}
	return room, nil
}

func (uc *RoomUseCase) Update(ctx context.Context, id string, req *domain.UpdateRoomRequest) (*domain.Room, error) {
	room, err := uc.roomRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	if req.RoomNumber != "" {
		room.RoomNumber = req.RoomNumber
	}
	if req.Floor != 0 {
		room.Floor = req.Floor
	}
	if req.Type != "" {
		room.Type = req.Type
	}
	if req.Status != "" {
		room.Status = req.Status
	}
	if req.RentPrice > 0 {
		room.RentPrice = req.RentPrice
	}
	room.Description = req.Description
	if err := uc.roomRepo.Update(ctx, room); err != nil {
		if errors.Is(err, domain.ErrDuplicate) {
			return nil, apierror.Conflict("ห้องหมายเลขนี้มีอยู่แล้ว")
		}
		return nil, apierror.Internal(err)
	}
	return room, nil
}

func (uc *RoomUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.roomRepo.FindByID(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.roomRepo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
