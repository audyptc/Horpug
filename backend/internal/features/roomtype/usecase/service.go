package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
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

// ActivityLogger records room type create/update/delete events for the audit
// trail. Failures to record are logged but never block the room type flow.
type ActivityLogger interface {
	Create(ctx context.Context, input activitylogusecase.CreateInput) (activitylogdomain.ActivityLog, error)
}

type Service struct {
	repo        Repository
	activityLog ActivityLogger
}

func New(repo Repository, activityLog ActivityLogger) *Service {
	return &Service{repo: repo, activityLog: activityLog}
}

// recordActivity is best-effort: a failure to write the audit trail must
// never fail the room type CRUD flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, entityID uuid.UUID, description string) {
	if s.activityLog == nil {
		return
	}
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "roomtype",
		EntityID:    &entityID,
		Description: description,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
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

	roomType, err := s.repo.Create(ctx, input)
	if err != nil {
		return roomtypedomain.RoomType{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", roomType.ID, fmt.Sprintf("Created room type: %s", roomType.Name))
	return roomType, nil
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

	roomType, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return roomtypedomain.RoomType{}, err
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", roomType.ID, fmt.Sprintf("Updated room type: %s", roomType.Name))
	return roomType, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	roomType, _ := s.repo.GetByID(ctx, id, requesterID)

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", id, fmt.Sprintf("Deleted room type: %s", roomType.Name))
	return nil
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
