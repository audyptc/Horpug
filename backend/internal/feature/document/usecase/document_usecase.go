package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/document/domain"

	"github.com/google/uuid"
)

type DocumentUseCase struct {
	repo domain.DocumentRepository
}

func NewDocumentUseCase(repo domain.DocumentRepository) *DocumentUseCase {
	return &DocumentUseCase{repo: repo}
}

func (uc *DocumentUseCase) List(ctx context.Context, limit, offset int) ([]*domain.DocumentDetail, int, error) {
	total, err := uc.repo.Count(ctx)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	list, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return list, total, nil
}

func (uc *DocumentUseCase) GetByID(ctx context.Context, id string) (*domain.DocumentDetail, error) {
	d, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return d, nil
}

func (uc *DocumentUseCase) Create(ctx context.Context, req *domain.CreateDocumentRequest) (*domain.DocumentDetail, error) {
	d := &domain.Document{
		ID:         uuid.New().String(),
		Title:      req.Title,
		Category:   req.Category,
		TenantID:   req.TenantID,
		FileURL:    req.FileURL,
		IssueDate:  req.IssueDate,
		ExpiryDate: req.ExpiryDate,
		Note:       req.Note,
	}
	if err := uc.repo.Create(ctx, d); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, d.ID)
}

func (uc *DocumentUseCase) Update(ctx context.Context, id string, req *domain.UpdateDocumentRequest) (*domain.DocumentDetail, error) {
	detail, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
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
	return uc.repo.FindByID(ctx, id)
}

func (uc *DocumentUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
