package usecase

import (
	"context"
	"strings"
	"time"

	announcementdomain "apihorpug/internal/features/announcement/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	DormitoryID *uuid.UUID
	IsPublished *bool
	DateFrom    *time.Time
	DateTo      *time.Time

	Search string
	// Columns narrows individual columns by substring, keyed by FilterColumns.
	// Search casts a wide OR across dormitory, title and content; these are
	// ANDed on top, so the two answer different questions and compose.
	Columns  map[string]string
	SortKey  string
	SortDesc bool
}

// FilterColumns whitelists the columns a caller may match a substring against,
// passed as f[<column>]=value. The repository resolves each key to the actual
// SQL it needs — this map exists purely so an unrecognised column is rejected
// rather than silently ignored.
var FilterColumns = map[string]string{
	"dormitory_name": "dormitory_name",
	"title":          "title",
}

// SortColumns whitelists the sort keys the API accepts. As with FilterColumns,
// the repository resolves each key to its actual ORDER BY expression.
var SortColumns = map[string]string{
	"dormitory_name": "dormitory_name",
	"title":          "title",
	"is_published":   "is_published",
	"published_date": "published_date",
}

const DefaultSortKey = "published_date"

type CreateInput struct {
	DormitoryID   uuid.UUID
	Title         string
	Content       string
	IsPublished   *bool
	PublishedDate time.Time
	CreatedBy     *uuid.UUID
}

type UpdateInput struct {
	Title         *string
	Content       *string
	IsPublished   *bool
	PublishedDate *time.Time
	UpdatedBy     *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]announcementdomain.Announcement, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (announcementdomain.Announcement, error)
	Create(ctx context.Context, input CreateInput) (announcementdomain.Announcement, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (announcementdomain.Announcement, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]announcementdomain.Announcement, int64, error) {
	filters.Search = strings.TrimSpace(filters.Search)
	if _, ok := SortColumns[filters.SortKey]; !ok {
		filters.SortKey = DefaultSortKey
	}

	columns := make(map[string]string, len(filters.Columns))
	for key, value := range filters.Columns {
		if _, ok := FilterColumns[key]; !ok {
			continue
		}
		if value = strings.TrimSpace(value); value != "" {
			columns[key] = value
		}
	}
	filters.Columns = columns

	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	announcements, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return announcements, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (announcementdomain.Announcement, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (announcementdomain.Announcement, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	if input.PublishedDate.IsZero() {
		input.PublishedDate = time.Now()
	}
	if input.IsPublished == nil {
		published := true
		input.IsPublished = &published
	}

	if input.DormitoryID == uuid.Nil || input.Title == "" {
		return announcementdomain.Announcement{}, announcementdomain.ErrRequiredAnnouncementData
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (announcementdomain.Announcement, error) {
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return announcementdomain.Announcement{}, announcementdomain.ErrRequiredAnnouncementData
		}
		input.Title = &title
	}
	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		input.Content = &content
	}

	return s.repo.Update(ctx, id, requesterID, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
