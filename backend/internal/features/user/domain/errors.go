package domain

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserDuplicate    = errors.New("username or email already exists")
	ErrRoleNotFound     = errors.New("role not found")
	ErrInvalidPassword  = errors.New("password cannot be empty")
	ErrInvalidUsername  = errors.New("username cannot be empty")
	ErrInvalidEmail     = errors.New("email cannot be empty")
	ErrRequiredUserData = errors.New("username, email and password are required")
	ErrUserProtected    = errors.New("user is protected and cannot be modified or deleted")
)
