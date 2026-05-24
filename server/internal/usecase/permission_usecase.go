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

func (uc *PermissionUseCase) List(ctx context.Context) ([]*domain.Permission, error) {
	return uc.permRepo.List(ctx)
}
