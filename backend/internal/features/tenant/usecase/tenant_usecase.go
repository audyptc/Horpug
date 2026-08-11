package usecase

import (
	"context"
	"errors"

	contractdomain "apigofiberhorpug/internal/features/contract/domain"
	"apigofiberhorpug/internal/features/tenant/domain"
	coredomain "apigofiberhorpug/internal/shared/domain"
	"apigofiberhorpug/internal/shared/http/apierror"

	"github.com/google/uuid"
)

type TenantUseCase struct {
	tenantRepo   domain.TenantRepository
	contractRepo contractdomain.ContractRepository
}

func NewTenantUseCase(tenantRepo domain.TenantRepository, contractRepo contractdomain.ContractRepository) *TenantUseCase {
	return &TenantUseCase{tenantRepo: tenantRepo, contractRepo: contractRepo}
}

func (uc *TenantUseCase) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.Tenant, int, error) {
	total, err := uc.tenantRepo.Count(ctx, dormitoryID)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	tenants, err := uc.tenantRepo.List(ctx, dormitoryID, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return tenants, total, nil
}

func (uc *TenantUseCase) GetByID(ctx context.Context, dormitoryID, id string) (*domain.Tenant, error) {
	t, err := uc.tenantRepo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return t, nil
}

func (uc *TenantUseCase) Create(ctx context.Context, dormitoryID string, req *domain.CreateTenantRequest, actorID string) (*domain.Tenant, error) {
	t := &domain.Tenant{
		ID:               uuid.New().String(),
		DormitoryID:      dormitoryID,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Phone:            req.Phone,
		IDCard:           req.IDCard,
		Email:            req.Email,
		EmergencyContact: req.EmergencyContact,
		Note:             req.Note,
		CreatedBy:        actorID,
		UpdatedBy:        actorID,
	}
	if err := uc.tenantRepo.Create(ctx, t); err != nil {
		return nil, apierror.Internal(err)
	}
	return t, nil
}

func (uc *TenantUseCase) Update(ctx context.Context, dormitoryID, id string, req *domain.UpdateTenantRequest, actorID string) (*domain.Tenant, error) {
	t, err := uc.tenantRepo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
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
	t.UpdatedBy = actorID
	if err := uc.tenantRepo.Update(ctx, t); err != nil {
		return nil, apierror.Internal(err)
	}
	return t, nil
}

func (uc *TenantUseCase) Delete(ctx context.Context, dormitoryID, id string) error {
	if _, err := uc.tenantRepo.FindByID(ctx, dormitoryID, id); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	hasContract, err := uc.contractRepo.HasActiveContractForTenant(ctx, id)
	if err != nil {
		return apierror.Internal(err)
	}
	if hasContract {
		return apierror.Conflict("ไม่สามารถลบผู้เช่าที่มีสัญญาเช่าที่ยังใช้งานอยู่ได้")
	}
	if err := uc.tenantRepo.Delete(ctx, dormitoryID, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
