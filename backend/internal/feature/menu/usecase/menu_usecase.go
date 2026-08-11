package usecase

import (
	"context"
	"errors"

	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/menu/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
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
		return nil, 0, apierror.Internal(err)
	}
	menus, err := uc.menuRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return menus, total, nil
}

func (uc *MenuUseCase) GetByID(ctx context.Context, id string) (*domain.Menu, error) {
	menu, err := uc.menuRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return menu, nil
}
