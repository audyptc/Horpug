package domain

import "context"

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
