package usecase

import (
	"context"

	"apigofiberhorpug/internal/domain"
)

type PermissionUseCase struct {
	permRepo domain.PermissionRepository
}

func NewPermissionUseCase(permRepo domain.PermissionRepository) *PermissionUseCase {
	return &PermissionUseCase{permRepo: permRepo}
}

func (uc *PermissionUseCase) List(ctx context.Context, limit, offset int) ([]*domain.Permission, int, error) {
	total, err := uc.permRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	perms, err := uc.permRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return perms, total, nil
}
