package domain

import "errors"

var (
	ErrMeterNotFound        = errors.New("meter reading not found")
	ErrRequiredMeterData    = errors.New("room_id and reading_date are required")
	ErrInvalidMeterUnits    = errors.New("previous_unit and current_unit must not be negative, and current_unit must not be less than previous_unit")
	ErrInvalidMeterPrice    = errors.New("price_per_unit must not be negative")
	ErrInvalidBillingMethod = errors.New("invalid billing method")
	ErrRequiredFlatAmount   = errors.New("flat_amount is required and must not be negative when billing_method is flat")
	ErrRoomNotFound         = errors.New("room not found")
	ErrMeterReadingExists   = errors.New("a meter reading for this room and date already exists")
)
