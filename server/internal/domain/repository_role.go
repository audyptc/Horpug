package domain

import "context"

type RoleRepository interface {
	FindByID(ctx context.Context, id string) (*Role, error)
	List(ctx context.Context) ([]*Role, error)
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id string) error
	AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error
	GetPermissions(ctx context.Context, roleID string) ([]Permission, error)
}