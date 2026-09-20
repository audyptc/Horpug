package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	announcementdomain "apihorpug/internal/features/announcement/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	DormitoryID *uuid.UUID
	IsPublished *bool
	Category    string
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
	"category":       "category",
	"is_published":   "is_published",
	"published_date": "published_date",
}

const DefaultSortKey = "published_date"

type CreateInput struct {
	DormitoryID   uuid.UUID
	Title         string
	Content       string
	Category      string
	IsPinned      bool
	IsPublished   *bool
	PublishedDate time.Time
	CreatedBy     *uuid.UUID
}

type UpdateInput struct {
	Title         *string
	Content       *string
	Category      *string
	IsPinned      *bool
	IsPublished   *bool
	PublishedDate *time.Time
	UpdatedBy     *uuid.UUID
}

type Repository interface {
	Summary(ctx context.Context, requesterID uuid.UUID) (announcementdomain.Summary, error)
	MarkRead(ctx context.Context, id, requesterID uuid.UUID) error
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]announcementdomain.Announcement, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (announcementdomain.Announcement, error)
	Create(ctx context.Context, input CreateInput) (announcementdomain.Announcement, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (announcementdomain.Announcement, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

// ActivityLogger records announcement create/update/delete events for the
// audit trail. Failures to record are logged but never block the announcement
// flow.
type ActivityLogger interface {
	Create(ctx context.Context, input activitylogusecase.CreateInput) (activitylogdomain.ActivityLog, error)
}

type Service struct {
	repo        Repository
	activityLog ActivityLogger
}

func New(repo Repository, activityLog ActivityLogger) *Service {
	return &Service{repo: repo, activityLog: activityLog}
}

// recordActivity is best-effort: a failure to write the audit trail must
// never fail the announcement flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, announcement announcementdomain.Announcement, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	var dormitoryRef *uuid.UUID
	if announcement.DormitoryID != uuid.Nil {
		dormitoryRef = &announcement.DormitoryID
	}
	entityID := announcement.ID
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "announcement",
		EntityID:    &entityID,
		DormitoryID: dormitoryRef,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func announcementStatus(isPublished bool) string {
	if isPublished {
		return "published"
	}
	return "draft"
}

// announcementUpdateDescription names the announcement and spells out the
// changes that matter to readers (going live, being hidden, pinned or moved to
// another category); a plain wording edit is just "Updated announcement: ...".
func announcementUpdateDescription(before, after announcementdomain.Announcement) string {
	description := fmt.Sprintf("Updated announcement: %s", after.Title)

	changes := make([]string, 0, 3)
	if before.IsPublished != after.IsPublished {
		changes = append(changes, fmt.Sprintf("status %s -> %s", announcementStatus(before.IsPublished), announcementStatus(after.IsPublished)))
	}
	if before.IsPinned != after.IsPinned {
		changes = append(changes, fmt.Sprintf("pinned %t -> %t", before.IsPinned, after.IsPinned))
	}
	if before.Category != after.Category {
		changes = append(changes, fmt.Sprintf("category %s -> %s", before.Category, after.Category))
	}
	if len(changes) > 0 {
		description += " (" + strings.Join(changes, ", ") + ")"
	}
	return description
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]announcementdomain.Announcement, int64, error) {
	filters.Search = strings.TrimSpace(filters.Search)
	filters.Category = strings.TrimSpace(filters.Category)
	if filters.Category != "" && !announcementdomain.ValidCategory(filters.Category) {
		return nil, 0, announcementdomain.ErrInvalidCategory
	}
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

// Summary reports the requester's unread count and whether they may manage
// announcements, for the sidebar badge and the list's action controls.
func (s *Service) Summary(ctx context.Context, requesterID uuid.UUID) (announcementdomain.Summary, error) {
	return s.repo.Summary(ctx, requesterID)
}

// MarkRead records that the requester has opened the announcement. Doing it
// twice is harmless.
func (s *Service) MarkRead(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.MarkRead(ctx, id, requesterID)
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (announcementdomain.Announcement, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (announcementdomain.Announcement, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	if input.PublishedDate.IsZero() {
		input.PublishedDate = time.Now()
	}
	if input.IsPublished == nil {
		published := true
		input.IsPublished = &published
	}

	input.Category = strings.TrimSpace(input.Category)
	if input.Category == "" {
		input.Category = announcementdomain.CategoryGeneral
	}
	if !announcementdomain.ValidCategory(input.Category) {
		return announcementdomain.Announcement{}, announcementdomain.ErrInvalidCategory
	}

	if input.DormitoryID == uuid.Nil || input.Title == "" {
		return announcementdomain.Announcement{}, announcementdomain.ErrRequiredAnnouncementData
	}

	announcement, err := s.repo.Create(ctx, input)
	if err != nil {
		return announcementdomain.Announcement{}, err
	}

	description := fmt.Sprintf("Created announcement: %s (%s, %s)", announcement.Title, announcement.Category, announcementStatus(announcement.IsPublished))
	s.recordActivity(ctx, input.CreatedBy, "CREATE", announcement, description, ipAddress)
	return announcement, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (announcementdomain.Announcement, error) {
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
	if input.Category != nil {
		category := strings.TrimSpace(*input.Category)
		if !announcementdomain.ValidCategory(category) {
			return announcementdomain.Announcement{}, announcementdomain.ErrInvalidCategory
		}
		input.Category = &category
	}

	// Read before the update so the log can show what changed. It also checks
	// the requester's access, so a missing or out-of-scope announcement fails
	// here as not found, same as the update itself would.
	before, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return announcementdomain.Announcement{}, err
	}

	announcement, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return announcementdomain.Announcement{}, err
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", announcement, announcementUpdateDescription(before, announcement), ipAddress)
	return announcement, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	// Read first: once deleted there is no title left to describe.
	announcement, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", announcement, fmt.Sprintf("Deleted announcement: %s", announcement.Title), ipAddress)
	return nil
}
