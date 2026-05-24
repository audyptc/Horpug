package domain

import "context"

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	AssignRoles(ctx context.Context, userID string, roleIDs []string) error
	GetRoles(ctx context.Context, userID string) ([]Role, error)
	GetPermissions(ctx context.Context, userID string) ([]string, error)
}