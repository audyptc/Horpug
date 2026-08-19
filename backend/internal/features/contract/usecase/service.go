package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	contractdomain "apihorpug/internal/features/contract/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	TenantID    *uuid.UUID
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
	Status      *contractdomain.ContractStatus
}

type CreateInput struct {
	TenantID     uuid.UUID
	RoomID       uuid.UUID
	StartDate    time.Time
	EndDate      *time.Time
	RentPrice    float64
	Deposit      float64
	NumOccupants int
	Note         string
	CreatedBy    *uuid.UUID
}

type UpdateInput struct {
	StartDate    *time.Time
	EndDate      *time.Time
	RentPrice    *float64
	Deposit      *float64
	NumOccupants *int
	Status       *contractdomain.ContractStatus
	Note         *string
	UpdatedBy    *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]contractdomain.Contract, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (contractdomain.Contract, error)
	Create(ctx context.Context, input CreateInput) (contractdomain.Contract, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (contractdomain.Contract, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

// ActivityLogger records contract create/update/delete events for the audit
// trail. Failures to record are logged but never block the contract flow.
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
// never fail the contract CRUD flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, entityID uuid.UUID, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "contract",
		EntityID:    &entityID,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]contractdomain.Contract, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	contracts, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return contracts, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (contractdomain.Contract, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (contractdomain.Contract, error) {
	input.Note = strings.TrimSpace(input.Note)

	if input.TenantID == uuid.Nil || input.RoomID == uuid.Nil || input.StartDate.IsZero() {
		return contractdomain.Contract{}, contractdomain.ErrRequiredContractData
	}
	if input.EndDate != nil && input.EndDate.Before(input.StartDate) {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractDates
	}
	if input.RentPrice < 0 || input.Deposit < 0 {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractAmount
	}
	if input.NumOccupants <= 0 {
		input.NumOccupants = 1
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (contractdomain.Contract, error) {
	if input.StartDate != nil && input.StartDate.IsZero() {
		return contractdomain.Contract{}, contractdomain.ErrRequiredContractData
	}
	if input.StartDate != nil && input.EndDate != nil && input.EndDate.Before(*input.StartDate) {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractDates
	}
	if input.RentPrice != nil && *input.RentPrice < 0 {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractAmount
	}
	if input.Deposit != nil && *input.Deposit < 0 {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractAmount
	}
	if input.NumOccupants != nil && *input.NumOccupants <= 0 {
		return contractdomain.Contract{}, contractdomain.ErrInvalidNumOccupants
	}
	if input.Status != nil && !input.Status.Valid() {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractStatus
	}
	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		input.Note = &note
	}

	return s.repo.Update(ctx, id, requesterID, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	contract, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return err
	}
	if contract.Status == contractdomain.ContractStatusActive {
		return contractdomain.ErrContractIsActive
	}

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", id,
		fmt.Sprintf("Deleted contract: %s - room %s", contract.TenantName, contract.RoomNumber), ipAddress)
	return nil
}
