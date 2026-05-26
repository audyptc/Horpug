package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"

	"github.com/google/uuid"
)

type TenantUseCase struct {
	tenantRepo domain.TenantRepository
}

func NewTenantUseCase(tenantRepo domain.TenantRepository) *TenantUseCase {
	return &TenantUseCase{tenantRepo: tenantRepo}
}

func (uc *TenantUseCase) List(ctx context.Context, limit, offset int) ([]*domain.Tenant, int, error) {
	total, err := uc.tenantRepo.Count(ctx)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	tenants, err := uc.tenantRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return tenants, total, nil
}

func (uc *TenantUseCase) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	t, err := uc.tenantRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return t, nil
}

func (uc *TenantUseCase) Create(ctx context.Context, req *domain.CreateTenantRequest) (*domain.Tenant, error) {
	t := &domain.Tenant{
		ID:               uuid.New().String(),
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Phone:            req.Phone,
		IDCard:           req.IDCard,
		Email:            req.Email,
		EmergencyContact: req.EmergencyContact,
		Note:             req.Note,
	}
	if err := uc.tenantRepo.Create(ctx, t); err != nil {
		return nil, apierror.Internal(err)
	}
	return t, nil
}

func (uc *TenantUseCase) Update(ctx context.Context, id string, req *domain.UpdateTenantRequest) (*domain.Tenant, error) {
	t, err := uc.tenantRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	if req.FirstName != "" {
		t.FirstName = req.FirstName
	}
	if req.LastName != "" {
		t.LastName = req.LastName
	}
	if req.Phone != "" {
		t.Phone = req.Phone
	}
	if req.IDCard != "" {
		t.IDCard = req.IDCard
	}
	t.Email = req.Email
	t.EmergencyContact = req.EmergencyContact
	t.Note = req.Note
	if err := uc.tenantRepo.Update(ctx, t); err != nil {
		return nil, apierror.Internal(err)
	}
	return t, nil
}

func (uc *TenantUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.tenantRepo.FindByID(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.tenantRepo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
