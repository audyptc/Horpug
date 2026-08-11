package usecase

import (
	"context"
	"errors"

	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/announcement/domain"
	"apigofiberhorpug/internal/shared/http/apierror"

	"github.com/google/uuid"
)

type AnnouncementUseCase struct {
	repo domain.AnnouncementRepository
}

func NewAnnouncementUseCase(repo domain.AnnouncementRepository) *AnnouncementUseCase {
	return &AnnouncementUseCase{repo: repo}
}

func (uc *AnnouncementUseCase) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.Announcement, int, error) {
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

func (uc *AnnouncementUseCase) GetByID(ctx context.Context, dormitoryID, id string) (*domain.Announcement, error) {
	a, err := uc.repo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return a, nil
}

func (uc *AnnouncementUseCase) Create(ctx context.Context, dormitoryID string, req *domain.CreateAnnouncementRequest) (*domain.Announcement, error) {
	a := &domain.Announcement{
		ID:          uuid.New().String(),
		DormitoryID: dormitoryID,
		Title:       req.Title,
		Content:     req.Content,
		Type:        req.Type,
		IsPinned:    req.IsPinned,
		PublishedAt: req.PublishedAt,
		ExpiredAt:   req.ExpiredAt,
	}
	if err := uc.repo.Create(ctx, a); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, dormitoryID, a.ID)
}

func (uc *AnnouncementUseCase) Update(ctx context.Context, dormitoryID, id string, req *domain.UpdateAnnouncementRequest) (*domain.Announcement, error) {
	a, err := uc.repo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	a.Title = req.Title
	a.Content = req.Content
	a.Type = req.Type
	a.IsPinned = req.IsPinned
	a.PublishedAt = req.PublishedAt
	a.ExpiredAt = req.ExpiredAt

	if err := uc.repo.Update(ctx, a); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, dormitoryID, id)
}

func (uc *AnnouncementUseCase) Delete(ctx context.Context, dormitoryID, id string) error {
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
