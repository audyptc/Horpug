package domain

import "errors"

var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrRequiredRoomData  = errors.New("room number, dormitory and room type are required")
	ErrInvalidRoomStatus = errors.New("invalid room status")
	ErrDormitoryNotFound = errors.New("dormitory not found")
	ErrRoomTypeNotFound  = errors.New("room type not found")
	ErrRoomNumberExists  = errors.New("room number already exists in this dormitory")
)
