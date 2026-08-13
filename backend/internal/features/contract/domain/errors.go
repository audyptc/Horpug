package domain

import "errors"

var (
	ErrContractNotFound      = errors.New("contract not found")
	ErrRequiredContractData  = errors.New("tenant_id, room_id and start_date are required")
	ErrInvalidContractStatus = errors.New("invalid contract status")
	ErrInvalidContractDates  = errors.New("end_date must not be before start_date")
	ErrInvalidContractAmount = errors.New("rent_price and deposit must not be negative")
	ErrInvalidNumOccupants   = errors.New("num_occupants must be at least 1")
	ErrTenantNotFound        = errors.New("tenant not found")
	ErrRoomNotFound          = errors.New("room not found")
	ErrRoomHasActiveContract = errors.New("room already has an active contract")
)
