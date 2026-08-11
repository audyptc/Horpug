package domain

import (
	"context"
	"time"

	permissiondomain "apigofiberhorpug/internal/features/permission/domain"
)

type Role struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	IsActive        bool                 `json:"is_active"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	MenuPermissions []RoleMenuPermission `json:"menu_permissions,omitempty"`
}

type RoleMenuPermission struct {
	MenuID      string                        `json:"menu_id"`
	MenuName    string                        `json:"menu_name"`
	Permissions []permissiondomain.Permission `json:"permissions"`
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type RoleMenuPermissionItem struct {
	MenuID        string   `json:"menu_id"`
	PermissionIDs []string `json:"permission_ids"`
}

type AssignPermissionsRequest struct {
	Items []RoleMenuPermissionItem `json:"items"`
}

type RoleRepository interface {
	FindByID(ctx context.Context, id string) (*Role, error)
	List(ctx context.Context, limit, offset int) ([]*Role, error)
	ListActive(ctx context.Context) ([]*Role, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id string) error
	AssignMenuPermissions(ctx context.Context, roleID string, items []RoleMenuPermissionItem) error
	GetMenuPermissions(ctx context.Context, roleID string) ([]RoleMenuPermission, error)
}
