package usecase

import (
	"context"

	"apigofiberhorpug/internal/domain"
)

type MenuUseCase struct {
	menuRepo domain.MenuRepository
}

func NewMenuUseCase(menuRepo domain.MenuRepository) *MenuUseCase {
	return &MenuUseCase{menuRepo: menuRepo}
}

func (uc *MenuUseCase) List(ctx context.Context) ([]*domain.Menu, error) {
	return uc.menuRepo.List(ctx)
}

func (uc *MenuUseCase) GetByID(ctx context.Context, id string) (*domain.Menu, error) {
	return uc.menuRepo.FindByID(ctx, id)
}
