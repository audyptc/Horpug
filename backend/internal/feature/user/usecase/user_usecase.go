package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/user/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCase struct {
	userRepo domain.UserRepository
	roleRepo coredomain.RoleRepository
}

func NewUserUseCase(userRepo domain.UserRepository, roleRepo coredomain.RoleRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepo, roleRepo: roleRepo}
}

func (uc *UserUseCase) List(ctx context.Context, limit, offset int) ([]*domain.User, int, error) {
	total, err := uc.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	users, err := uc.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	for _, u := range users {
		u.Role, err = uc.userRepo.GetRole(ctx, u.ID)
		if err != nil {
			return nil, 0, apierror.Internal(err)
		}
	}
	return users, total, nil
}

func (uc *UserUseCase) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	user.Role, err = uc.userRepo.GetRole(ctx, id)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	return user, nil
}

func (uc *UserUseCase) Create(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	hashed, err := hashPassword(req.Password)
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
		return nil, apierror.Internal(err)
	}
	user.Password = ""
	return user, nil
}

func (uc *UserUseCase) Update(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.Password != "" {
		hashed, err := hashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hashed
	}
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, apierror.Internal(err)
	}
	user.Role, err = uc.userRepo.GetRole(ctx, id)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	return user, nil
}

func (uc *UserUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.userRepo.FindByID(ctx, id); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.userRepo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}

func (uc *UserUseCase) AssignRole(ctx context.Context, userID string, req *domain.AssignRoleRequest) error {
	if _, err := uc.userRepo.FindByID(ctx, userID); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if _, err := uc.roleRepo.FindByID(ctx, req.RoleID); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound("role not found")
		}
		return apierror.Internal(err)
	}
	if err := uc.userRepo.AssignRole(ctx, userID, req.RoleID); err != nil {
		return apierror.Internal(err)
	}
	return nil
}

func hashPassword(raw string) (string, error) {
	if len(raw) < 8 {
		return "", apierror.BadRequest("password must be at least 8 characters")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", apierror.Internal(err)
	}
	return string(hashed), nil
}
