package usecase

import (
	"context"

	menudomain "apihorpug/internal/features/menu/domain"
)

type Repository interface {
	List(ctx context.Context) ([]menudomain.Menu, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]menudomain.Menu, error) {
	return s.repo.List(ctx)
}
