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

func (uc *MenuUseCase) List(ctx context.Context, limit, offset int) ([]*domain.Menu, int, error) {
	total, err := uc.menuRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	menus, err := uc.menuRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return menus, total, nil
}

func (uc *MenuUseCase) GetByID(ctx context.Context, id string) (*domain.Menu, error) {
	return uc.menuRepo.FindByID(ctx, id)
}
