package usecase

import (
	"context"
	"strings"

	permissiondomain "apihorpug/internal/features/permission/domain"
)

type CreateInput struct {
	Name        string
	Description string
}

type Repository interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context, limit, offset int) ([]permissiondomain.Permission, error)
	Create(ctx context.Context, input CreateInput) (permissiondomain.Permission, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]permissiondomain.Permission, int64, error) {
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	permissions, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (permissiondomain.Permission, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Name == "" {
		return permissiondomain.Permission{}, permissiondomain.ErrPermissionNameRequired
	}
	if !permissiondomain.Action(input.Name).Valid() {
		return permissiondomain.Permission{}, permissiondomain.ErrPermissionNameInvalid
	}

	return s.repo.Create(ctx, input)
}
