package domain

import (
	"context"
	"time"
)

type Dormitory struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateDormitoryRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type UpdateDormitoryRequest struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	IsActive *bool  `json:"is_active"`
}

type AssignDormitoriesRequest struct {
	DormitoryIDs []string `json:"dormitory_ids"`
}

type DormitoryRepository interface {
	FindByID(ctx context.Context, id string) (*Dormitory, error)
	List(ctx context.Context, limit, offset int) ([]*Dormitory, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, dormitory *Dormitory) error
	Update(ctx context.Context, dormitory *Dormitory) error
	Delete(ctx context.Context, id string) error

	ListForUser(ctx context.Context, userID string) ([]*Dormitory, error)
	HasAccess(ctx context.Context, userID, dormitoryID string) (bool, error)
	SetUserDormitories(ctx context.Context, userID string, dormitoryIDs []string) error
}
