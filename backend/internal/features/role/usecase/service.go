package usecase

import (
	"context"
	"strings"

	roledomain "apihorpug/internal/features/role/domain"

	"github.com/google/uuid"
)

type MenuPermissionInput struct {
	MenuID        uuid.UUID
	PermissionIDs []uuid.UUID
}

type CreateInput struct {
	Name                string
	Description         string
	IsActive            bool
	FullDormitoryAccess bool
	DormitoryIDs        []uuid.UUID
	MenuPermissions     []MenuPermissionInput
	CreatedBy           *uuid.UUID
}

type UpdateInput struct {
	Name                *string
	Description         *string
	IsActive            *bool
	FullDormitoryAccess *bool
	DormitoryIDs        *[]uuid.UUID
	MenuPermissions     *[]MenuPermissionInput
	UpdatedBy           *uuid.UUID
}

type DeletionCheck struct {
	CanDelete   bool  `json:"can_delete"`
	UserCount   int64 `json:"user_count"`
	IsProtected bool  `json:"is_protected"`
}

type Repository interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context, limit, offset int) ([]roledomain.Role, error)
	ListActive(ctx context.Context, search string, limit int) ([]roledomain.Role, error)
	GetByID(ctx context.Context, id uuid.UUID) (roledomain.Role, error)
	CountUsers(ctx context.Context, id uuid.UUID) (int64, error)
	Create(ctx context.Context, input CreateInput) (roledomain.Role, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (roledomain.Role, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]roledomain.Role, int64, error) {
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	roles, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (s *Service) ListActive(ctx context.Context, search string, limit int) ([]roledomain.Role, error) {
	return s.repo.ListActive(ctx, strings.TrimSpace(search), limit)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (roledomain.Role, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (roledomain.Role, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (roledomain.Role, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return roledomain.Role{}, err
	}
	if role.IsProtected {
		return roledomain.Role{}, roledomain.ErrRoleProtected
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		input.Name = &name
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		input.Description = &description
	}
	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role.IsProtected {
		return roledomain.ErrRoleProtected
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) CheckDeletion(ctx context.Context, id uuid.UUID) (DeletionCheck, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return DeletionCheck{}, err
	}

	userCount, err := s.repo.CountUsers(ctx, id)
	if err != nil {
		return DeletionCheck{}, err
	}

	return DeletionCheck{
		CanDelete:   !role.IsProtected && userCount == 0,
		UserCount:   userCount,
		IsProtected: role.IsProtected,
	}, nil
}
