package domain

import "errors"

var (
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrRequiredTenantData = errors.New("first_name and last_name are required")
	ErrTenantIDCardExists = errors.New("id card already exists")
	ErrTenantHasContracts = errors.New("tenant has contracts and cannot be deleted")
	ErrInvalidLineToken   = errors.New("invalid or expired LINE id token")

	// ErrLineAccountAlreadyLinked means the LINE account behind a LIFF id
	// token is already linked to a different tenant — most often because
	// whoever opened this tenant's linking link was, at the time, logged
	// into LINE as an account already linked elsewhere (e.g. staff testing
	// the flow with their own LINE account across several tenants).
	ErrLineAccountAlreadyLinked = errors.New("this LINE account is already linked to another tenant")
)
