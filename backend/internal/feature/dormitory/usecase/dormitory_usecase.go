package usecase

import (
	"context"
	"errors"
	"slices"
	"strings"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/dormitory/domain"

	"github.com/google/uuid"
)

type DormitoryUseCase struct {
	dormitoryRepo domain.DormitoryRepository
}

func NewDormitoryUseCase(dormitoryRepo domain.DormitoryRepository) *DormitoryUseCase {
	return &DormitoryUseCase{dormitoryRepo: dormitoryRepo}
}

func (uc *DormitoryUseCase) List(ctx context.Context, limit, offset int) ([]*domain.Dormitory, int, error) {
	total, err := uc.dormitoryRepo.Count(ctx)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	dormitories, err := uc.dormitoryRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return dormitories, total, nil
}

func (uc *DormitoryUseCase) GetByID(ctx context.Context, id string) (*domain.Dormitory, error) {
	dormitory, err := uc.dormitoryRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return dormitory, nil
}

func (uc *DormitoryUseCase) Create(ctx context.Context, req *domain.CreateDormitoryRequest) (*domain.Dormitory, error) {
	dormitory := &domain.Dormitory{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Address:  req.Address,
		IsActive: true,
	}
	if err := uc.dormitoryRepo.Create(ctx, dormitory); err != nil {
		return nil, apierror.Internal(err)
	}
	return dormitory, nil
}

func (uc *DormitoryUseCase) Update(ctx context.Context, id string, req *domain.UpdateDormitoryRequest) (*domain.Dormitory, error) {
	dormitory, err := uc.dormitoryRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	if req.Name != "" {
		dormitory.Name = req.Name
	}
	dormitory.Address = req.Address
	if req.IsActive != nil {
		dormitory.IsActive = *req.IsActive
	}
	if err := uc.dormitoryRepo.Update(ctx, dormitory); err != nil {
		return nil, apierror.Internal(err)
	}
	return dormitory, nil
}

func (uc *DormitoryUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.dormitoryRepo.FindByID(ctx, id); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.dormitoryRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, coredomain.ErrDuplicate) {
			return apierror.Conflict("ไม่สามารถลบหอพักที่ยังมีห้อง/ผู้เช่าอยู่ได้")
		}
		return apierror.Internal(err)
	}
	return nil
}

// ListAccessible returns every dormitory the given user may operate against.
// Admins can access every dormitory; everyone else is limited to their assignments.
func (uc *DormitoryUseCase) ListAccessible(ctx context.Context, userID, roleName string) ([]*domain.Dormitory, error) {
	filterActive := func(dormitories []*domain.Dormitory) []*domain.Dormitory {
		return slices.DeleteFunc(dormitories, func(d *domain.Dormitory) bool {
			return d == nil || !d.IsActive
		})
	}

	if strings.EqualFold(roleName, "admin") {
		dormitories, err := uc.dormitoryRepo.List(ctx, 1000, 0)
		if err != nil {
			return nil, apierror.Internal(err)
		}
		return filterActive(dormitories), nil
	}
	dormitories, err := uc.dormitoryRepo.ListForUser(ctx, userID)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	return filterActive(dormitories), nil
}

// ListForUser returns the raw set of dormitories a user is explicitly assigned to,
// with no admin bypass — used to show real state in an assignment-editing UI.
func (uc *DormitoryUseCase) ListForUser(ctx context.Context, userID string) ([]*domain.Dormitory, error) {
	dormitories, err := uc.dormitoryRepo.ListForUser(ctx, userID)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	return dormitories, nil
}

// CheckAccess reports whether the given user may operate against dormitoryID.
func (uc *DormitoryUseCase) CheckAccess(ctx context.Context, userID, roleName, dormitoryID string) (bool, error) {
	if strings.EqualFold(roleName, "admin") {
		return true, nil
	}
	ok, err := uc.dormitoryRepo.HasAccess(ctx, userID, dormitoryID)
	if err != nil {
		return false, apierror.Internal(err)
	}
	return ok, nil
}

// AssignDormitoriesToUser replaces the full set of dormitories a user may access.
func (uc *DormitoryUseCase) AssignDormitoriesToUser(ctx context.Context, userID string, req *domain.AssignDormitoriesRequest) error {
	if err := uc.dormitoryRepo.SetUserDormitories(ctx, userID, req.DormitoryIDs); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
