package domain

import (
	"time"

	"github.com/google/uuid"
)

// Announcement is a notice posted to tenants of a dormitory (e.g. maintenance
// schedule, rule change, event).
type Announcement struct {
	ID            uuid.UUID  `json:"id"`
	DormitoryID   uuid.UUID  `json:"dormitory_id"`
	DormitoryName string     `json:"dormitory_name,omitempty"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	IsPublished   bool       `json:"is_published"`
	PublishedDate time.Time  `json:"published_date"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy     *uuid.UUID `json:"updated_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
