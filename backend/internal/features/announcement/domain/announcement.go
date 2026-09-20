package domain

import (
	"time"

	"github.com/google/uuid"
)

// Categories is the closed set of announcement types, so the list can be
// filtered and badged consistently. "urgent" is the one tenants should notice.
const (
	CategoryGeneral     = "general"
	CategoryUrgent      = "urgent"
	CategoryBilling     = "billing"
	CategoryMaintenance = "maintenance"
	CategoryEvent       = "event"
)

var Categories = []string{
	CategoryGeneral,
	CategoryUrgent,
	CategoryBilling,
	CategoryMaintenance,
	CategoryEvent,
}

func ValidCategory(category string) bool {
	for _, valid := range Categories {
		if valid == category {
			return true
		}
	}
	return false
}

// Announcement is a notice posted to tenants of a dormitory (e.g. maintenance
// schedule, rule change, event).
type Announcement struct {
	ID            uuid.UUID `json:"id"`
	DormitoryID   uuid.UUID `json:"dormitory_id"`
	DormitoryName string    `json:"dormitory_name,omitempty"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Category      string    `json:"category"`
	// IsPinned keeps the announcement at the top of the default listing.
	IsPinned      bool      `json:"is_pinned"`
	IsPublished   bool      `json:"is_published"`
	PublishedDate time.Time `json:"published_date"`
	// IsRead is whether the requesting user has opened this announcement; it
	// is per user, not a property of the announcement itself.
	IsRead    bool       `json:"is_read"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Summary is what the shell needs without loading the list: how many
// announcements the user hasn't opened yet, and whether they may manage them
// (so tenants aren't shown create/edit/delete controls that would only 403).
type Summary struct {
	UnreadCount int64 `json:"unread_count"`
	CanManage   bool  `json:"can_manage"`
}
