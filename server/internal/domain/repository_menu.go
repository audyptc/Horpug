package domain

import "context"

type MenuRepository interface {
	FindByID(ctx context.Context, id string) (*Menu, error)
	List(ctx context.Context, limit, offset int) ([]*Menu, error)
	Count(ctx context.Context) (int, error)
}
