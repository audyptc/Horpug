package domain

import "errors"

var (
	ErrDormitoryNotFound     = errors.New("dormitory not found")
	ErrRequiredDormitoryData = errors.New("name is required")
	ErrManagerNotFound       = errors.New("one or more users not found")
	ErrDormitoryHasRooms     = errors.New("dormitory has rooms and cannot be deleted")
	ErrInvalidPromptPayID    = errors.New("promptpay_id must be a 10-digit mobile number, 13-digit national/tax id or 15-digit e-wallet id")
)
