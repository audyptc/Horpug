package domain

import "errors"

var (
	ErrPermissionNameRequired = errors.New("name is required")
	ErrPermissionNameExists   = errors.New("permission name already exists")
	ErrPermissionNameInvalid  = errors.New("name must be one of the predefined actions")
)
