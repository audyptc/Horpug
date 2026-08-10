package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/maintenancerequest/domain"
	roomdomain "apigofiberhorpug/internal/feature/room/domain"

	"github.com/google/uuid"
)

type MaintenanceRequestUseCase struct {
	repo     domain.MaintenanceRequestRepository
	roomRepo roomdomain.RoomRepository
}

func NewMaintenanceRequestUseCase(repo domain.MaintenanceRequestRepository, roomRepo roomdomain.RoomRepository) *MaintenanceRequestUseCase {
	return &MaintenanceRequestUseCase{repo: repo, roomRepo: roomRepo}
}

func (uc *MaintenanceRequestUseCase) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.MaintenanceRequestDetail, int, error) {
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

func (uc *MaintenanceRequestUseCase) GetByID(ctx context.Context, dormitoryID, id string) (*domain.MaintenanceRequestDetail, error) {
	m, err := uc.repo.FindDetailByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return m, nil
}

func (uc *MaintenanceRequestUseCase) validateRoom(ctx context.Context, dormitoryID, roomID string) error {
	if _, err := uc.roomRepo.FindByID(ctx, dormitoryID, roomID); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound("room not found")
		}
		return apierror.Internal(err)
	}
	return nil
}

func (uc *MaintenanceRequestUseCase) Create(ctx context.Context, dormitoryID string, req *domain.CreateMaintenanceRequestRequest) (*domain.MaintenanceRequestDetail, error) {
	if err := uc.validateRoom(ctx, dormitoryID, req.RoomID); err != nil {
		return nil, err
	}
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
	return uc.repo.FindDetailByID(ctx, dormitoryID, m.ID)
}

func (uc *MaintenanceRequestUseCase) Update(ctx context.Context, dormitoryID, id string, req *domain.UpdateMaintenanceRequestRequest) (*domain.MaintenanceRequestDetail, error) {
	m, err := uc.repo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	if req.RoomID != m.RoomID {
		if err := uc.validateRoom(ctx, dormitoryID, req.RoomID); err != nil {
			return nil, err
		}
	}

	m.RoomID = req.RoomID
	m.Title = req.Title
	m.Description = req.Description
	m.Status = req.Status
	m.Priority = req.Priority
	m.ReportedDate = req.ReportedDate
	m.ResolvedDate = req.ResolvedDate
	m.Note = req.Note

	if err := uc.repo.Update(ctx, dormitoryID, m); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindDetailByID(ctx, dormitoryID, id)
}

func (uc *MaintenanceRequestUseCase) Delete(ctx context.Context, dormitoryID, id string) error {
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
