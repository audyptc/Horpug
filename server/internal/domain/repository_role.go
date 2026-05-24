package domain

import "context"

type RoleRepository interface {
	FindByID(ctx context.Context, id string) (*Role, error)
	List(ctx context.Context, limit, offset int) ([]*Role, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id string) error
	AssignMenuPermissions(ctx context.Context, roleID string, items []RoleMenuPermissionItem) error
	GetPermissions(ctx context.Context, roleID string) ([]Permission, error)
}
