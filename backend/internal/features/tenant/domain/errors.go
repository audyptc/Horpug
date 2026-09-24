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

	// ErrTenantLineAlreadyLinked means this tenant already has a different
	// LINE account linked. The public linking endpoint only needs a tenant id,
	// so letting it overwrite an existing link would let anyone holding the
	// link redirect that tenant's invoices to their own LINE. Staff must
	// unlink first (DELETE /tenants/:id/line) to allow a new link.
	ErrTenantLineAlreadyLinked = errors.New("this tenant already has a LINE account linked")
)
