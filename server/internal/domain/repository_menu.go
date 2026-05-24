package domain

import "context"

type MenuRepository interface {
	FindByID(ctx context.Context, id string) (*Menu, error)
	List(ctx context.Context) ([]*Menu, error)
}
