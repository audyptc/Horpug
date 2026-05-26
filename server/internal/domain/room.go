package domain

import (
	"context"
	"time"
)

type Room struct {
	ID          string    `json:"id"`
	RoomNumber  string    `json:"room_number"`
	Floor       int       `json:"floor"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	RentPrice   float64   `json:"rent_price"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRoomRequest struct {
	RoomNumber  string  `json:"room_number"`
	Floor       int     `json:"floor"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	RentPrice   float64 `json:"rent_price"`
	Description string  `json:"description"`
}

type UpdateRoomRequest struct {
	RoomNumber  string  `json:"room_number"`
	Floor       int     `json:"floor"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	RentPrice   float64 `json:"rent_price"`
	Description string  `json:"description"`
}

type RoomRepository interface {
	FindByID(ctx context.Context, id string) (*Room, error)
	List(ctx context.Context, limit, offset int) ([]*Room, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, room *Room) error
	Update(ctx context.Context, room *Room) error
	Delete(ctx context.Context, id string) error
}
