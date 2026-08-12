package usecase

import (
	"context"
	"strings"

	userdomain "apihorpug/internal/features/user/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserPermissionItem struct {
	MenuID         uuid.UUID `json:"menu_id"`
	MenuName       string    `json:"menu_name"`
	MenuPath       string    `json:"menu_path"`
	PermissionID   uuid.UUID `json:"permission_id"`
	PermissionName string    `json:"permission_name"`
}

type CreateInput struct {
	Username  string
	Email     string
	Password  string
	RoleID    uuid.UUID
	IsActive  bool
	CreatedBy *uuid.UUID
}

type UpdateInput struct {
	Username  *string
	Email     *string
	Password  *string
	RoleID    *uuid.UUID
	IsActive  *bool
	UpdatedBy *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context, limit, offset int) ([]userdomain.User, error)
	ListActive(ctx context.Context, search string, limit int) ([]userdomain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (userdomain.User, error)
	GetPermissions(ctx context.Context, id uuid.UUID) ([]UserPermissionItem, error)
	Create(ctx context.Context, input CreateInput, hashedPassword string) (userdomain.User, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput, hashedPassword *string) (userdomain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]userdomain.User, int64, error) {
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *Service) ListActive(ctx context.Context, search string, limit int) ([]userdomain.User, error) {
	return s.repo.ListActive(ctx, strings.TrimSpace(search), limit)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (userdomain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetPermissions(ctx context.Context, id uuid.UUID) ([]UserPermissionItem, error) {
	return s.repo.GetPermissions(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (userdomain.User, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Username == "" || input.Email == "" || strings.TrimSpace(input.Password) == "" {
		return userdomain.User{}, userdomain.ErrRequiredUserData
	}
	if input.RoleID == uuid.Nil {
		return userdomain.User{}, userdomain.ErrRoleNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return userdomain.User{}, err
	}

	return s.repo.Create(ctx, input, string(hashedPassword))
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (userdomain.User, error) {
	if input.Username != nil {
		username := strings.TrimSpace(*input.Username)
		if username == "" {
			return userdomain.User{}, userdomain.ErrInvalidUsername
		}
		input.Username = &username
	}
	if input.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*input.Email))
		if email == "" {
			return userdomain.User{}, userdomain.ErrInvalidEmail
		}
		input.Email = &email
	}
	if input.RoleID != nil && *input.RoleID == uuid.Nil {
		return userdomain.User{}, userdomain.ErrRoleNotFound
	}

	var hashedPassword *string
	if input.Password != nil {
		password := strings.TrimSpace(*input.Password)
		if password == "" {
			return userdomain.User{}, userdomain.ErrInvalidPassword
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return userdomain.User{}, err
		}
		hashedPasswordValue := string(hashed)
		hashedPassword = &hashedPasswordValue
	}

	return s.repo.Update(ctx, id, input, hashedPassword)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
