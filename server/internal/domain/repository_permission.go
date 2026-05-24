package domain

import "context"

type PermissionRepository interface {
	List(ctx context.Context) ([]*Permission, error)
	FindByID(ctx context.Context, id string) (*Permission, error)
}
