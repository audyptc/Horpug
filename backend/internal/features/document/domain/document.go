package domain

import (
	"time"

	"github.com/google/uuid"
)

type DocumentCategory string

const (
	DocumentCategoryContract DocumentCategory = "contract"
	DocumentCategoryIDCard   DocumentCategory = "id_card"
	DocumentCategoryReceipt  DocumentCategory = "receipt"
	DocumentCategoryOther    DocumentCategory = "other"
)

func (c DocumentCategory) Valid() bool {
	switch c {
	case DocumentCategoryContract, DocumentCategoryIDCard, DocumentCategoryReceipt, DocumentCategoryOther:
		return true
	}
	return false
}

// Document is a general file attachment stored for a dormitory, optionally
// tied to a specific tenant or room it concerns (e.g. a signed contract scan
// or a copy of a tenant's ID card).
type Document struct {
	ID            uuid.UUID        `json:"id"`
	DormitoryID   uuid.UUID        `json:"dormitory_id"`
	DormitoryName string           `json:"dormitory_name,omitempty"`
	TenantID      *uuid.UUID       `json:"tenant_id,omitempty"`
	TenantName    *string          `json:"tenant_name,omitempty"`
	RoomID        *uuid.UUID       `json:"room_id,omitempty"`
	RoomNumber    *string          `json:"room_number,omitempty"`
	Name          string           `json:"name"`
	Category      DocumentCategory `json:"category"`
	FileURL       string           `json:"file_url"`
	// FilePath is the storage key of an uploaded file; it is never exposed —
	// clients fetch the bytes through the authenticated file endpoint.
	FilePath     string     `json:"-"`
	FileName     string     `json:"file_name,omitempty"`
	FileMime     string     `json:"file_mime,omitempty"`
	FileSize     int64      `json:"file_size,omitempty"`
	HasFile      bool       `json:"has_file"`
	UploadedDate time.Time  `json:"uploaded_date"`
	Note         string     `json:"note"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy    *uuid.UUID `json:"updated_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
