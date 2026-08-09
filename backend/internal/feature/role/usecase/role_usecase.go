package usecase

import (
	"context"
	"errors"
	"strings"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	menudomain "apigofiberhorpug/internal/feature/menu/domain"
	permissiondomain "apigofiberhorpug/internal/feature/permission/domain"
	"apigofiberhorpug/internal/feature/role/domain"

	"github.com/google/uuid"
)

type RoleUseCase struct {
	roleRepo domain.RoleRepository
	permRepo permissiondomain.PermissionRepository
	menuRepo menudomain.MenuRepository
}

func NewRoleUseCase(roleRepo domain.RoleRepository, permRepo permissiondomain.PermissionRepository, menuRepo menudomain.MenuRepository) *RoleUseCase {
	return &RoleUseCase{roleRepo: roleRepo, permRepo: permRepo, menuRepo: menuRepo}
}

func (uc *RoleUseCase) List(ctx context.Context, limit, offset int) ([]*domain.Role, int, error) {
	total, err := uc.roleRepo.Count(ctx)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	roles, err := uc.roleRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	for _, r := range roles {
		r.MenuPermissions, err = uc.roleRepo.GetMenuPermissions(ctx, r.ID)
		if err != nil {
			return nil, 0, apierror.Internal(err)
		}
	}
	return roles, total, nil
}

func (uc *RoleUseCase) ListActive(ctx context.Context) ([]*domain.Role, error) {
	roles, err := uc.roleRepo.ListActive(ctx)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	return roles, nil
}

func (uc *RoleUseCase) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	role, err := uc.roleRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	role.MenuPermissions, err = uc.roleRepo.GetMenuPermissions(ctx, id)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	return role, nil
}

func (uc *RoleUseCase) Create(ctx context.Context, req *domain.CreateRoleRequest) (*domain.Role, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	role := &domain.Role{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}
	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, apierror.Internal(err)
	}
	role.MenuPermissions = []domain.RoleMenuPermission{}
	return role, nil
}

func (uc *RoleUseCase) Update(ctx context.Context, id string, req *domain.UpdateRoleRequest) (*domain.Role, error) {
	role, err := uc.roleRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	if strings.EqualFold(role.Name, "admin") {
		return nil, apierror.Forbidden("cannot modify admin role")
	}
	if req.Name != "" {
		role.Name = req.Name
	}
	role.Description = req.Description
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}
	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, apierror.Internal(err)
	}
	role.MenuPermissions, err = uc.roleRepo.GetMenuPermissions(ctx, id)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	return role, nil
}

func (uc *RoleUseCase) Delete(ctx context.Context, id string) error {
	role, err := uc.roleRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if strings.EqualFold(role.Name, "admin") {
		return apierror.Forbidden("cannot delete admin role")
	}
	if err := uc.roleRepo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}

func (uc *RoleUseCase) AssignPermissions(ctx context.Context, roleID string, req *domain.AssignPermissionsRequest) error {
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if strings.EqualFold(role.Name, "admin") {
		return apierror.Forbidden("cannot modify permissions for admin role")
	}
	for _, item := range req.Items {
		if _, err := uc.menuRepo.FindByID(ctx, item.MenuID); err != nil {
			if errors.Is(err, coredomain.ErrNotFound) {
				return apierror.NotFound("menu not found: " + item.MenuID)
			}
			return apierror.Internal(err)
		}
		for _, permID := range item.PermissionIDs {
			if _, err := uc.permRepo.FindByID(ctx, permID); err != nil {
				if errors.Is(err, coredomain.ErrNotFound) {
					return apierror.NotFound("permission not found: " + permID)
				}
				return apierror.Internal(err)
			}
		}
	}
	if err := uc.roleRepo.AssignMenuPermissions(ctx, roleID, req.Items); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
