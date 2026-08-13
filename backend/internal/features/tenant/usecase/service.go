package usecase

import (
	"context"
	"strings"

	tenantdomain "apihorpug/internal/features/tenant/domain"

	"github.com/google/uuid"
)

type CreateInput struct {
	FirstName        string
	LastName         string
	Phone            string
	IDCard           string
	Email            string
	EmergencyContact string
	Note             string
	IsActive         bool
	CreatedBy        *uuid.UUID
}

type UpdateInput struct {
	FirstName        *string
	LastName         *string
	Phone            *string
	IDCard           *string
	Email            *string
	EmergencyContact *string
	Note             *string
	IsActive         *bool
	UpdatedBy        *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context, limit, offset int) ([]tenantdomain.Tenant, error)
	ListActive(ctx context.Context, search string, limit int) ([]tenantdomain.Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (tenantdomain.Tenant, error)
	Create(ctx context.Context, input CreateInput) (tenantdomain.Tenant, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (tenantdomain.Tenant, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]tenantdomain.Tenant, int64, error) {
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	tenants, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}

func (s *Service) ListActive(ctx context.Context, search string, limit int) ([]tenantdomain.Tenant, error) {
	return s.repo.ListActive(ctx, strings.TrimSpace(search), limit)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (tenantdomain.Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (tenantdomain.Tenant, error) {
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.IDCard = strings.TrimSpace(input.IDCard)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.EmergencyContact = strings.TrimSpace(input.EmergencyContact)
	input.Note = strings.TrimSpace(input.Note)

	if input.FirstName == "" || input.LastName == "" {
		return tenantdomain.Tenant{}, tenantdomain.ErrRequiredTenantData
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (tenantdomain.Tenant, error) {
	if input.FirstName != nil {
		firstName := strings.TrimSpace(*input.FirstName)
		if firstName == "" {
			return tenantdomain.Tenant{}, tenantdomain.ErrRequiredTenantData
		}
		input.FirstName = &firstName
	}
	if input.LastName != nil {
		lastName := strings.TrimSpace(*input.LastName)
		if lastName == "" {
			return tenantdomain.Tenant{}, tenantdomain.ErrRequiredTenantData
		}
		input.LastName = &lastName
	}
	if input.Phone != nil {
		phone := strings.TrimSpace(*input.Phone)
		input.Phone = &phone
	}
	if input.IDCard != nil {
		idCard := strings.TrimSpace(*input.IDCard)
		input.IDCard = &idCard
	}
	if input.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*input.Email))
		input.Email = &email
	}
	if input.EmergencyContact != nil {
		emergencyContact := strings.TrimSpace(*input.EmergencyContact)
		input.EmergencyContact = &emergencyContact
	}
	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		input.Note = &note
	}

	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
