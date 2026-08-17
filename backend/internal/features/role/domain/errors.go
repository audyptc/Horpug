package domain

import "errors"

var (
	ErrRoleNotFound      = errors.New("role not found")
	ErrReferenceNotFound = errors.New("one or more menus or permissions not found")
	ErrRoleNameExists    = errors.New("role name already exists")
	ErrRoleInUse         = errors.New("role is being used by users")
	ErrRoleProtected     = errors.New("role is protected and cannot be modified or deleted")
)
