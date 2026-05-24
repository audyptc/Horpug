package usecase

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/domain"

	"github.com/google/uuid"
)

type RoleUseCase struct {
	roleRepo domain.RoleRepository
	permRepo domain.PermissionRepository
}

func NewRoleUseCase(roleRepo domain.RoleRepository, permRepo domain.PermissionRepository) *RoleUseCase {
	return &RoleUseCase{roleRepo: roleRepo, permRepo: permRepo}
}

func (uc *RoleUseCase) List(ctx context.Context, limit, offset int) ([]*domain.Role, int, error) {
	total, err := uc.roleRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	roles, err := uc.roleRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	for _, r := range roles {
		r.Permissions, err = uc.roleRepo.GetPermissions(ctx, r.ID)
		if err != nil {
			return nil, 0, err
		}
	}
	return roles, total, nil
}

func (uc *RoleUseCase) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	role, err := uc.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	role.Permissions, err = uc.roleRepo.GetPermissions(ctx, id)
	return role, err
}

func (uc *RoleUseCase) Create(ctx context.Context, req *domain.CreateRoleRequest) (*domain.Role, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	role := &domain.Role{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
	}
	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("could not create role: %v", err)
	}
	role.Permissions = []domain.Permission{}
	return role, nil
}

func (uc *RoleUseCase) Update(ctx context.Context, id string, req *domain.UpdateRoleRequest) (*domain.Role, error) {
	role, err := uc.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		role.Name = req.Name
	}
	role.Description = req.Description
	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}
	role.Permissions, err = uc.roleRepo.GetPermissions(ctx, id)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (uc *RoleUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.roleRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return uc.roleRepo.Delete(ctx, id)
}

func (uc *RoleUseCase) AssignPermissions(ctx context.Context, roleID string, req *domain.AssignPermissionsRequest) error {
	if _, err := uc.roleRepo.FindByID(ctx, roleID); err != nil {
		return err
	}
	for _, permID := range req.PermissionIDs {
		if _, err := uc.permRepo.FindByID(ctx, permID); err != nil {
			return fmt.Errorf("permission %s not found", permID)
		}
	}
	return uc.roleRepo.AssignPermissions(ctx, roleID, req.PermissionIDs)
}
