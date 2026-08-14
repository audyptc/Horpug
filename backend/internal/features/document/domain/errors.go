package domain

import "errors"

var (
	ErrDocumentNotFound        = errors.New("document not found")
	ErrRequiredDocumentData    = errors.New("dormitory_id, name and file_url are required")
	ErrInvalidDocumentCategory = errors.New("invalid document category")
	ErrDormitoryNotFound       = errors.New("dormitory not found")
	ErrTenantNotFound          = errors.New("tenant not found")
	ErrRoomNotFound            = errors.New("room not found")
)
