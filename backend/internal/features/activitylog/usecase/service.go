package usecase

import (
	"context"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"

	"github.com/google/uuid"
)

type CreateInput struct {
	UserID      *uuid.UUID
	Action      string
	EntityType  string
	EntityID    *uuid.UUID
	DormitoryID *uuid.UUID
	Description string
	IPAddress   string
}

// ListFilter narrows and orders an activity log listing.
type ListFilter struct {
	Search string
	// Columns narrows individual columns by substring, keyed by FilterColumns.
	// Search casts a wide OR across user, action, entity type, description and
	// IP; these are ANDed on top, so the two answer different questions and
	// compose.
	Columns  map[string]string
	UserID   *uuid.UUID
	EntityID *uuid.UUID
	// DormitoryID narrows to one dormitory. It is ANDed with the requester's
	// own dormitory scope, so it can only narrow what they may already see.
	DormitoryID *uuid.UUID
	DateFrom    *time.Time
	DateTo      *time.Time
	SortKey     string
	SortDesc    bool
	Limit       int
	Offset      int
}

// FilterColumns whitelists the columns a caller may match a substring
// against. Same rule as SortColumns: the column name is interpolated into SQL
// rather than bound, so nothing outside this map may reach the query.
var FilterColumns = map[string]string{
	"username":    "COALESCE(u.username, '')",
	"action":      "al.action",
	"entity_type": "al.entity_type",
	"description": "al.description",
	"ip_address":  "al.ip_address",
}

// SortColumns maps the sort keys the API accepts onto the columns they order
// by. A column name can't be passed to Postgres as a bind parameter, so it is
// interpolated into the query — every value that reaches ORDER BY must come
// from this map and never straight from the request.
var SortColumns = map[string]string{
	"created_at":  "al.created_at",
	"username":    "COALESCE(u.username, '')",
	"action":      "al.action",
	"entity_type": "al.entity_type",
	"description": "al.description",
	"ip_address":  "al.ip_address",
}

const DefaultSortKey = "created_at"

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filter ListFilter) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filter ListFilter) ([]activitylogdomain.ActivityLog, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (activitylogdomain.ActivityLog, error)
	Create(ctx context.Context, input CreateInput) (activitylogdomain.ActivityLog, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns every log for roles with full dormitory access, otherwise only
// logs tied to dormitories the requester manages.
func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filter ListFilter) ([]activitylogdomain.ActivityLog, int64, error) {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	total, err := s.repo.Count(ctx, requesterID, filter)
	if err != nil {
		return nil, 0, err
	}

	logs, err := s.repo.List(ctx, requesterID, filter)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (activitylogdomain.ActivityLog, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (activitylogdomain.ActivityLog, error) {
	input.Action = strings.TrimSpace(input.Action)
	input.EntityType = strings.TrimSpace(input.EntityType)
	input.Description = strings.TrimSpace(input.Description)
	return s.repo.Create(ctx, input)
}
