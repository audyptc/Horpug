package usecase

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCase struct {
	userRepo domain.UserRepository
	roleRepo domain.RoleRepository
}

func NewUserUseCase(userRepo domain.UserRepository, roleRepo domain.RoleRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepo, roleRepo: roleRepo}
}

func (uc *UserUseCase) List(ctx context.Context, limit, offset int) ([]*domain.User, int, error) {
	total, err := uc.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	users, err := uc.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	for _, u := range users {
		u.Roles, err = uc.userRepo.GetRoles(ctx, u.ID)
		if err != nil {
			return nil, 0, err
		}
	}
	return users, total, nil
}

func (uc *UserUseCase) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Roles, err = uc.userRepo.GetRoles(ctx, id)
	return user, err
}

func (uc *UserUseCase) Create(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	if req.FullName == "" || req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("full_name, email and password are required")
	}
	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:       uuid.New().String(),
		FullName: req.FullName,
		Email:    req.Email,
		Password: string(hashed),
		IsActive: true,
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("could not create user: %v", err)
	}
	user.Password = ""
	user.Roles = []domain.Role{}
	return user, nil
}

func (uc *UserUseCase) Update(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	user.Roles, _ = uc.userRepo.GetRoles(ctx, id)
	return user, nil
}

func (uc *UserUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.userRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return uc.userRepo.Delete(ctx, id)
}

func (uc *UserUseCase) AssignRoles(ctx context.Context, userID string, req *domain.AssignRolesRequest) error {
	if _, err := uc.userRepo.FindByID(ctx, userID); err != nil {
		return err
	}
	for _, roleID := range req.RoleIDs {
		if _, err := uc.roleRepo.FindByID(ctx, roleID); err != nil {
			return fmt.Errorf("role %s not found", roleID)
		}
	}
	return uc.userRepo.AssignRoles(ctx, userID, req.RoleIDs)
}
