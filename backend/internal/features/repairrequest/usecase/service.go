package usecase

import (
	"context"
	"strings"
	"time"

	repairrequestdomain "apihorpug/internal/features/repairrequest/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
	TenantID    *uuid.UUID
	Category    *repairrequestdomain.RepairCategory
	Status      *repairrequestdomain.RepairStatus
}

type CreateInput struct {
	RoomID       uuid.UUID
	TenantID     *uuid.UUID
	Category     repairrequestdomain.RepairCategory
	Description  string
	Status       repairrequestdomain.RepairStatus
	ReportedDate time.Time
	CreatedBy    *uuid.UUID
}

type UpdateInput struct {
	TenantID     *uuid.UUID
	Category     *repairrequestdomain.RepairCategory
	Description  *string
	Status       *repairrequestdomain.RepairStatus
	ReportedDate *time.Time
	UpdatedBy    *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]repairrequestdomain.RepairRequest, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (repairrequestdomain.RepairRequest, error)
	Create(ctx context.Context, input CreateInput) (repairrequestdomain.RepairRequest, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (repairrequestdomain.RepairRequest, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]repairrequestdomain.RepairRequest, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	requests, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (repairrequestdomain.RepairRequest, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (repairrequestdomain.RepairRequest, error) {
	input.Description = strings.TrimSpace(input.Description)
	if input.Category == "" {
		input.Category = repairrequestdomain.RepairCategoryOther
	}
	if input.Status == "" {
		input.Status = repairrequestdomain.RepairStatusPending
	}

	if input.RoomID == uuid.Nil || input.Description == "" || input.ReportedDate.IsZero() {
		return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrRequiredRepairRequestData
	}
	if !input.Category.Valid() {
		return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrInvalidCategory
	}
	if !input.Status.Valid() {
		return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrInvalidStatus
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (repairrequestdomain.RepairRequest, error) {
	if input.Category != nil && !input.Category.Valid() {
		return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrInvalidCategory
	}
	if input.Status != nil && !input.Status.Valid() {
		return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrInvalidStatus
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		if description == "" {
			return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrRequiredRepairRequestData
		}
		input.Description = &description
	}

	return s.repo.Update(ctx, id, requesterID, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
