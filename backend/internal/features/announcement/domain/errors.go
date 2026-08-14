package domain

import "errors"

var (
	ErrAnnouncementNotFound     = errors.New("announcement not found")
	ErrRequiredAnnouncementData = errors.New("dormitory_id and title are required")
	ErrDormitoryNotFound        = errors.New("dormitory not found")
)
