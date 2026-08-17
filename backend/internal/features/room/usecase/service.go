package usecase

import (
	"context"
	"strings"

	roomdomain "apihorpug/internal/features/room/domain"

	"github.com/google/uuid"
)

type CreateInput struct {
	DormitoryID uuid.UUID
	RoomTypeID  uuid.UUID
	RoomNumber  string
	Floor       int
	Status      roomdomain.RoomStatus
	IsActive    bool
	CreatedBy   *uuid.UUID
}

type UpdateInput struct {
	RoomTypeID *uuid.UUID
	RoomNumber *string
	Floor      *int
	Status     *roomdomain.RoomStatus
	IsActive   *bool
	UpdatedBy  *uuid.UUID
}

type DeletionCheck struct {
	CanDelete     bool  `json:"can_delete"`
	ContractCount int64 `json:"contract_count"`
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, limit, offset int) ([]roomdomain.Room, error)
	ListActive(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, search string, limit int) ([]roomdomain.Room, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (roomdomain.Room, error)
	CountContracts(ctx context.Context, id uuid.UUID) (int64, error)
	Create(ctx context.Context, input CreateInput) (roomdomain.Room, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (roomdomain.Room, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, limit, offset int) ([]roomdomain.Room, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, dormitoryID)
	if err != nil {
		return nil, 0, err
	}

	rooms, err := s.repo.List(ctx, requesterID, dormitoryID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return rooms, total, nil
}

func (s *Service) ListActive(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, search string, limit int) ([]roomdomain.Room, error) {
	return s.repo.ListActive(ctx, requesterID, dormitoryID, strings.TrimSpace(search), limit)
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (roomdomain.Room, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (roomdomain.Room, error) {
	input.RoomNumber = strings.TrimSpace(input.RoomNumber)
	if input.RoomNumber == "" || input.DormitoryID == uuid.Nil || input.RoomTypeID == uuid.Nil {
		return roomdomain.Room{}, roomdomain.ErrRequiredRoomData
	}
	if input.Status == "" {
		input.Status = roomdomain.RoomStatusAvailable
	}
	if !input.Status.Valid() {
		return roomdomain.Room{}, roomdomain.ErrInvalidRoomStatus
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (roomdomain.Room, error) {
	if input.RoomNumber != nil {
		roomNumber := strings.TrimSpace(*input.RoomNumber)
		if roomNumber == "" {
			return roomdomain.Room{}, roomdomain.ErrRequiredRoomData
		}
		input.RoomNumber = &roomNumber
	}
	if input.Status != nil && !input.Status.Valid() {
		return roomdomain.Room{}, roomdomain.ErrInvalidRoomStatus
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

	contractCount, err := s.repo.CountContracts(ctx, id)
	if err != nil {
		return DeletionCheck{}, err
	}

	return DeletionCheck{CanDelete: contractCount == 0, ContractCount: contractCount}, nil
}
