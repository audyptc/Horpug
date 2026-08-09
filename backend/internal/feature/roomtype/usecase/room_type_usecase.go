package usecase

import (
	"context"
	"errors"
	"strings"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/roomtype/domain"
)

type RoomTypeUseCase struct {
	repo domain.RoomTypeRepository
}

func NewRoomTypeUseCase(repo domain.RoomTypeRepository) *RoomTypeUseCase {
	return &RoomTypeUseCase{repo: repo}
}

func (uc *RoomTypeUseCase) List(ctx context.Context) ([]*domain.RoomType, error) {
	types, err := uc.repo.List(ctx)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	return types, nil
}

func (uc *RoomTypeUseCase) GetByID(ctx context.Context, id string) (*domain.RoomType, error) {
	rt, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return rt, nil
}

func (uc *RoomTypeUseCase) Create(ctx context.Context, req *domain.CreateRoomTypeRequest) (*domain.RoomType, error) {
	id := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(req.ID), " ", "_"))
	if id == "" {
		return nil, apierror.BadRequest("id is required")
	}
	if req.Name == "" {
		return nil, apierror.BadRequest("name is required")
	}
	rt := &domain.RoomType{
		ID:        id,
		Name:      req.Name,
		SortOrder: req.SortOrder,
	}
	if err := uc.repo.Create(ctx, rt); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, rt.ID)
}

func (uc *RoomTypeUseCase) Update(ctx context.Context, id string, req *domain.UpdateRoomTypeRequest) (*domain.RoomType, error) {
	rt, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	if req.Name != "" {
		rt.Name = req.Name
	}
	rt.SortOrder = req.SortOrder
	if err := uc.repo.Update(ctx, rt); err != nil {
		return nil, apierror.Internal(err)
	}
	return rt, nil
}

func (uc *RoomTypeUseCase) Delete(ctx context.Context, id string) error {
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
