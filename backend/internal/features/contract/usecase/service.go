package usecase

import (
	"context"
	"strings"
	"time"

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

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
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

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
