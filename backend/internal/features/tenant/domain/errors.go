package domain

import "errors"

var (
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrRequiredTenantData = errors.New("first_name and last_name are required")
	ErrTenantIDCardExists = errors.New("id card already exists")
)
