package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/room/domain"

	"github.com/google/uuid"
)

type RoomUseCase struct {
	roomRepo     domain.RoomRepository
	contractRepo coredomain.ContractRepository
}

func NewRoomUseCase(roomRepo domain.RoomRepository, contractRepo coredomain.ContractRepository) *RoomUseCase {
	return &RoomUseCase{roomRepo: roomRepo, contractRepo: contractRepo}
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
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return room, nil
}

func (uc *RoomUseCase) Create(ctx context.Context, req *domain.CreateRoomRequest, actorID string) (*domain.Room, error) {
	room := &domain.Room{
		ID:          uuid.New().String(),
		RoomNumber:  req.RoomNumber,
		Floor:       req.Floor,
		Type:        req.Type,
		Status:      req.Status,
		RentPrice:   req.RentPrice,
		Description: req.Description,
		CreatedBy:   actorID,
		UpdatedBy:   actorID,
	}
	if room.Type == "" {
		room.Type = "standard"
	}
	if room.Status == "" {
		room.Status = "available"
	}
	if err := uc.roomRepo.Create(ctx, room); err != nil {
		if errors.Is(err, coredomain.ErrDuplicate) {
			return nil, apierror.Conflict("ห้องหมายเลขนี้มีอยู่แล้ว")
		}
		return nil, apierror.Internal(err)
	}
	return room, nil
}

func (uc *RoomUseCase) Update(ctx context.Context, id string, req *domain.UpdateRoomRequest, actorID string) (*domain.Room, error) {
	room, err := uc.roomRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
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
	room.UpdatedBy = actorID
	if err := uc.roomRepo.Update(ctx, room); err != nil {
		if errors.Is(err, coredomain.ErrDuplicate) {
			return nil, apierror.Conflict("ห้องหมายเลขนี้มีอยู่แล้ว")
		}
		return nil, apierror.Internal(err)
	}
	return room, nil
}

func (uc *RoomUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.roomRepo.FindByID(ctx, id); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	hasContract, err := uc.contractRepo.HasActiveContractForRoom(ctx, id)
	if err != nil {
		return apierror.Internal(err)
	}
	if hasContract {
		return apierror.Conflict("ไม่สามารถลบห้องที่มีสัญญาเช่าที่ยังใช้งานอยู่ได้")
	}
	if err := uc.roomRepo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
