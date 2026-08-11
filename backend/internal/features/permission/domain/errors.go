package domain

import "errors"

var (
	ErrPermissionNameRequired = errors.New("name is required")
	ErrPermissionNameExists   = errors.New("permission name already exists")
)
