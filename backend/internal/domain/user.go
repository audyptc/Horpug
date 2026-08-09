package domain

import (
	"context"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Role      *Role     `json:"role,omitempty"`
}

type CreateUserRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name"`
	Password string `json:"password"`
	IsActive *bool  `json:"is_active"`
}

type AssignRoleRequest struct {
	RoleID string `json:"role_id"`
}

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	AssignRole(ctx context.Context, userID string, roleID string) error
	GetRole(ctx context.Context, userID string) (*Role, error)
	GetPermissions(ctx context.Context, userID string) ([]string, error)
}
