package usecase

import (
	"context"
	"strings"

	roomtypedomain "apihorpug/internal/features/roomtype/domain"

	"github.com/google/uuid"
)

type CreateInput struct {
	DormitoryID uuid.UUID
	Name        string
	Description string
	Price       float64
	IsActive    bool
	CreatedBy   *uuid.UUID
}

type UpdateInput struct {
	Name        *string
	Description *string
	Price       *float64
	IsActive    *bool
	UpdatedBy   *uuid.UUID
}

type DeletionCheck struct {
	CanDelete bool  `json:"can_delete"`
	RoomCount int64 `json:"room_count"`
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, limit, offset int) ([]roomtypedomain.RoomType, error)
	ListActive(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, search string, limit int) ([]roomtypedomain.RoomType, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (roomtypedomain.RoomType, error)
	CountRooms(ctx context.Context, id uuid.UUID) (int64, error)
	Create(ctx context.Context, input CreateInput) (roomtypedomain.RoomType, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (roomtypedomain.RoomType, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, limit, offset int) ([]roomtypedomain.RoomType, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, dormitoryID)
	if err != nil {
		return nil, 0, err
	}

	roomTypes, err := s.repo.List(ctx, requesterID, dormitoryID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return roomTypes, total, nil
}

func (s *Service) ListActive(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, search string, limit int) ([]roomtypedomain.RoomType, error) {
	return s.repo.ListActive(ctx, requesterID, dormitoryID, strings.TrimSpace(search), limit)
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (roomtypedomain.RoomType, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (roomtypedomain.RoomType, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Name == "" || input.DormitoryID == uuid.Nil {
		return roomtypedomain.RoomType{}, roomtypedomain.ErrRequiredRoomTypeData
	}
	if input.Price < 0 {
		return roomtypedomain.RoomType{}, roomtypedomain.ErrInvalidRoomTypePrice
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (roomtypedomain.RoomType, error) {
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return roomtypedomain.RoomType{}, roomtypedomain.ErrRequiredRoomTypeData
		}
		input.Name = &name
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		input.Description = &description
	}
	if input.Price != nil && *input.Price < 0 {
		return roomtypedomain.RoomType{}, roomtypedomain.ErrInvalidRoomTypePrice
	}

	return s.repo.Update(ctx, id, requesterID, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}

func (s *Service) CheckDeletion(ctx context.Context, id, requesterID uuid.UUID) (DeletionCheck, error) {
	if _, err := s.repo.GetByID(ctx, id, requesterID); err != nil {
		return DeletionCheck{}, err
	}

	roomCount, err := s.repo.CountRooms(ctx, id)
	if err != nil {
		return DeletionCheck{}, err
	}

	return DeletionCheck{CanDelete: roomCount == 0, RoomCount: roomCount}, nil
}
