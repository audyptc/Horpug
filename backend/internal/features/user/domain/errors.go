package domain

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUserDuplicate = errors.New("username or email already exists")
	ErrRoleNotFound  = errors.New("role not found")
	// ErrRoleNotAssignable means the role exists but grants more than the
	// requester holds themselves (full dormitory access, a menu permission or
	// a dormitory they lack), so they may not hand it to a user.
	ErrRoleNotAssignable = errors.New("role cannot be assigned by this user")
	ErrInvalidPassword   = errors.New("password cannot be empty")
	ErrInvalidUsername   = errors.New("username cannot be empty")
	ErrInvalidEmail      = errors.New("email cannot be empty")
	ErrRequiredUserData  = errors.New("username, email and password are required")
	ErrUserProtected     = errors.New("user is protected and cannot be modified or deleted")
)
