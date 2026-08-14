package domain

import "errors"

var (
	ErrParcelNotFound      = errors.New("parcel not found")
	ErrRequiredParcelData  = errors.New("tenant_id and received_date are required")
	ErrInvalidParcelStatus = errors.New("invalid parcel status")
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrRoomNotFound        = errors.New("room not found")
)
