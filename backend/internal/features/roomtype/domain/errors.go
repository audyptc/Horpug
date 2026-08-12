package domain

import "errors"

var (
	ErrRoomTypeNotFound     = errors.New("room type not found")
	ErrRequiredRoomTypeData = errors.New("name and dormitory are required")
	ErrDormitoryNotFound    = errors.New("dormitory not found")
	ErrRoomTypeNameExists   = errors.New("room type name already exists in this dormitory")
)
