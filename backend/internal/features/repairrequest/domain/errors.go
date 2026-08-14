package domain

import "errors"

var (
	ErrRepairRequestNotFound     = errors.New("repair request not found")
	ErrRequiredRepairRequestData = errors.New("room_id, description and reported_date are required")
	ErrInvalidCategory           = errors.New("invalid repair category")
	ErrInvalidStatus             = errors.New("invalid repair status")
	ErrRoomNotFound              = errors.New("room not found")
	ErrTenantNotFound            = errors.New("tenant not found")
)
