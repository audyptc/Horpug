package domain

import (
	"context"
	"time"
)

type AnnouncementType string

const (
	AnnouncementTypeGeneral     AnnouncementType = "general"
	AnnouncementTypeMaintenance AnnouncementType = "maintenance"
	AnnouncementTypePayment     AnnouncementType = "payment"
	AnnouncementTypeEmergency   AnnouncementType = "emergency"
)

type Announcement struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Type        AnnouncementType `json:"type"`
	IsPinned    bool             `json:"is_pinned"`
	PublishedAt time.Time        `json:"published_at"`
	ExpiredAt   *time.Time       `json:"expired_at"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type CreateAnnouncementRequest struct {
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Type        AnnouncementType `json:"type"`
	IsPinned    bool             `json:"is_pinned"`
	PublishedAt time.Time        `json:"published_at"`
	ExpiredAt   *time.Time       `json:"expired_at"`
}

type UpdateAnnouncementRequest struct {
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Type        AnnouncementType `json:"type"`
	IsPinned    bool             `json:"is_pinned"`
	PublishedAt time.Time        `json:"published_at"`
	ExpiredAt   *time.Time       `json:"expired_at"`
}

type AnnouncementRepository interface {
	FindByID(ctx context.Context, id string) (*Announcement, error)
	List(ctx context.Context, limit, offset int) ([]*Announcement, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, a *Announcement) error
	Update(ctx context.Context, a *Announcement) error
	Delete(ctx context.Context, id string) error
}
