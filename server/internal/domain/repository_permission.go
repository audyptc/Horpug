package domain

import "context"

type PermissionRepository interface {
	List(ctx context.Context, limit, offset int) ([]*Permission, error)
	Count(ctx context.Context) (int, error)
	FindByID(ctx context.Context, id string) (*Permission, error)
}
