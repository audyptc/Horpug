package domain

import (
	"context"
	"time"
)

type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PermissionRepository interface {
	List(ctx context.Context, limit, offset int) ([]*Permission, error)
	Count(ctx context.Context) (int, error)
	FindByID(ctx context.Context, id string) (*Permission, error)
}
