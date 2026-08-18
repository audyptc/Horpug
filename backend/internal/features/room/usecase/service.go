package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
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

// ActivityLogger records room create/update/delete events for the audit
// trail. Failures to record are logged but never block the room flow.
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
// never fail the room CRUD flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, entityID uuid.UUID, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "room",
		EntityID:    &entityID,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
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

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (roomdomain.Room, error) {
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

	room, err := s.repo.Create(ctx, input)
	if err != nil {
		return roomdomain.Room{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", room.ID, fmt.Sprintf("Created room: %s", room.RoomNumber), ipAddress)
	return room, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (roomdomain.Room, error) {
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

	room, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return roomdomain.Room{}, err
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", room.ID, fmt.Sprintf("Updated room: %s", room.RoomNumber), ipAddress)
	return room, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	room, _ := s.repo.GetByID(ctx, id, requesterID)

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", id, fmt.Sprintf("Deleted room: %s", room.RoomNumber), ipAddress)
	return nil
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
