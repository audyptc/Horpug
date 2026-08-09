package domain

import (
	"context"
	"time"
)

type RoomType struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRoomTypeRequest struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type UpdateRoomTypeRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type RoomTypeRepository interface {
	List(ctx context.Context) ([]*RoomType, error)
	FindByID(ctx context.Context, id string) (*RoomType, error)
	Create(ctx context.Context, rt *RoomType) error
	Update(ctx context.Context, rt *RoomType) error
	Delete(ctx context.Context, id string) error
}
