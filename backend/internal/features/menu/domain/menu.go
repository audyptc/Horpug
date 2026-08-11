package domain

import (
	"context"
	"time"
)

//go:generate go run ../../../../cmd/modelgen -input menu.go -type Menu -out ../../../database/migrations/001_create_menus.sql

// Menu represents a navigation item stored in the menus table.
type Menu struct {
	ID        string    `json:"id" db:"uuid,primarykey"`
	Name      string    `json:"name" db:"text,notnull"`
	Path      string    `json:"path,omitempty" db:"text,nullable"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"timestamptz,notnull,default:NOW()"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"timestamptz,notnull,default:NOW()"`
}

type MenuRepository interface {
	FindByID(ctx context.Context, id string) (*Menu, error)
	List(ctx context.Context, limit, offset int) ([]*Menu, error)
	Count(ctx context.Context) (int, error)
}
