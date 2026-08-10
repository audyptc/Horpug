package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/document/domain"
	tenantdomain "apigofiberhorpug/internal/feature/tenant/domain"

	"github.com/google/uuid"
)

type DocumentUseCase struct {
	repo       domain.DocumentRepository
	tenantRepo tenantdomain.TenantRepository
}

func NewDocumentUseCase(repo domain.DocumentRepository, tenantRepo tenantdomain.TenantRepository) *DocumentUseCase {
	return &DocumentUseCase{repo: repo, tenantRepo: tenantRepo}
}

func (uc *DocumentUseCase) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.DocumentDetail, int, error) {
	total, err := uc.repo.Count(ctx, dormitoryID)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	list, err := uc.repo.List(ctx, dormitoryID, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return list, total, nil
}

func (uc *DocumentUseCase) GetByID(ctx context.Context, dormitoryID, id string) (*domain.DocumentDetail, error) {
	d, err := uc.repo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return d, nil
}

func (uc *DocumentUseCase) validateTenant(ctx context.Context, dormitoryID string, tenantID *string) error {
	if tenantID == nil || *tenantID == "" {
		return nil
	}
	if _, err := uc.tenantRepo.FindByID(ctx, dormitoryID, *tenantID); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound("tenant not found")
		}
		return apierror.Internal(err)
	}
	return nil
}

func (uc *DocumentUseCase) Create(ctx context.Context, dormitoryID string, req *domain.CreateDocumentRequest) (*domain.DocumentDetail, error) {
	if err := uc.validateTenant(ctx, dormitoryID, req.TenantID); err != nil {
		return nil, err
	}
	d := &domain.Document{
		ID:          uuid.New().String(),
		DormitoryID: dormitoryID,
		Title:       req.Title,
		Category:    req.Category,
		TenantID:    req.TenantID,
		FileURL:     req.FileURL,
		IssueDate:   req.IssueDate,
		ExpiryDate:  req.ExpiryDate,
		Note:        req.Note,
	}
	if err := uc.repo.Create(ctx, d); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, dormitoryID, d.ID)
}

func (uc *DocumentUseCase) Update(ctx context.Context, dormitoryID, id string, req *domain.UpdateDocumentRequest) (*domain.DocumentDetail, error) {
	detail, err := uc.repo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	if err := uc.validateTenant(ctx, dormitoryID, req.TenantID); err != nil {
		return nil, err
	}

	d := &detail.Document
	d.Title = req.Title
	d.Category = req.Category
	d.TenantID = req.TenantID
	d.FileURL = req.FileURL
	d.IssueDate = req.IssueDate
	d.ExpiryDate = req.ExpiryDate
	d.Note = req.Note

	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, dormitoryID, id)
}

func (uc *DocumentUseCase) Delete(ctx context.Context, dormitoryID, id string) error {
	if _, err := uc.repo.FindByID(ctx, dormitoryID, id); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.repo.Delete(ctx, dormitoryID, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
