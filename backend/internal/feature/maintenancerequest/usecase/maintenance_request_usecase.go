package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/maintenancerequest/domain"

	"github.com/google/uuid"
)

type MaintenanceRequestUseCase struct {
	repo domain.MaintenanceRequestRepository
}

func NewMaintenanceRequestUseCase(repo domain.MaintenanceRequestRepository) *MaintenanceRequestUseCase {
	return &MaintenanceRequestUseCase{repo: repo}
}

func (uc *MaintenanceRequestUseCase) List(ctx context.Context, limit, offset int) ([]*domain.MaintenanceRequestDetail, int, error) {
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

func (uc *MaintenanceRequestUseCase) GetByID(ctx context.Context, id string) (*domain.MaintenanceRequestDetail, error) {
	m, err := uc.repo.FindDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return m, nil
}

func (uc *MaintenanceRequestUseCase) Create(ctx context.Context, req *domain.CreateMaintenanceRequestRequest) (*domain.MaintenanceRequestDetail, error) {
	m := &domain.MaintenanceRequest{
		ID:           uuid.New().String(),
		RoomID:       req.RoomID,
		Title:        req.Title,
		Description:  req.Description,
		Status:       req.Status,
		Priority:     req.Priority,
		ReportedDate: req.ReportedDate,
		ResolvedDate: req.ResolvedDate,
		Note:         req.Note,
	}
	if err := uc.repo.Create(ctx, m); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindDetailByID(ctx, m.ID)
}

func (uc *MaintenanceRequestUseCase) Update(ctx context.Context, id string, req *domain.UpdateMaintenanceRequestRequest) (*domain.MaintenanceRequestDetail, error) {
	m, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	m.RoomID = req.RoomID
	m.Title = req.Title
	m.Description = req.Description
	m.Status = req.Status
	m.Priority = req.Priority
	m.ReportedDate = req.ReportedDate
	m.ResolvedDate = req.ResolvedDate
	m.Note = req.Note

	if err := uc.repo.Update(ctx, m); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindDetailByID(ctx, id)
}

func (uc *MaintenanceRequestUseCase) Delete(ctx context.Context, id string) error {
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
