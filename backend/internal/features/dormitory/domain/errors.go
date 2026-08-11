package domain

import "errors"

var (
	ErrDormitoryNotFound     = errors.New("dormitory not found")
	ErrRequiredDormitoryData = errors.New("name is required")
	ErrManagerNotFound       = errors.New("one or more users not found")
)
