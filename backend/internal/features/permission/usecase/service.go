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
	List(ctx context.Context) ([]permissiondomain.Permission, error)
	Create(ctx context.Context, input CreateInput) (permissiondomain.Permission, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]permissiondomain.Permission, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (permissiondomain.Permission, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Name == "" {
		return permissiondomain.Permission{}, permissiondomain.ErrPermissionNameRequired
	}

	return s.repo.Create(ctx, input)
}
