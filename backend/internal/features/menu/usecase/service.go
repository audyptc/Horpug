package usecase

import (
	"context"

	menudomain "apihorpug/internal/features/menu/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context, limit, offset int) ([]menudomain.Menu, error)
	ListForRole(ctx context.Context, roleID uuid.UUID, action string) ([]menudomain.Menu, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]menudomain.Menu, int64, error) {
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	menus, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return menus, total, nil
}

// ListForRole returns the menus a role has read access to, i.e. the menus
// that should appear in navigation for a user with that role.
func (s *Service) ListForRole(ctx context.Context, roleID uuid.UUID) ([]menudomain.Menu, error) {
	return s.repo.ListForRole(ctx, roleID, "read")
}
