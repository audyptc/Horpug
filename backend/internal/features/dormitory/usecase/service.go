package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	dormdomain "apihorpug/internal/features/dormitory/domain"

	"github.com/google/uuid"
)

type CreateInput struct {
	Name        string
	Address     string
	Phone       string
	Description string
	IsActive    bool
	ManagerIDs  []uuid.UUID
	CreatedBy   *uuid.UUID
}

type UpdateInput struct {
	Name        *string
	Address     *string
	Phone       *string
	Description *string
	IsActive    *bool
	ManagerIDs  *[]uuid.UUID
	UpdatedBy   *uuid.UUID
}

type DeletionCheck struct {
	CanDelete bool  `json:"can_delete"`
	RoomCount int64 `json:"room_count"`
}

// ListFilters narrows and orders a dormitory listing. IsActive is nil-able so
// asking for inactive dormitories stays distinct from not filtering on status
// at all.
type ListFilters struct {
	Search string
	// Columns narrows individual columns by substring, keyed by FilterColumns.
	// Search casts a wide OR across every text column; these are ANDed on top,
	// so the two answer different questions and compose.
	Columns  map[string]string
	IsActive *bool
	SortKey  string
	SortDesc bool
}

// FilterColumns whitelists the columns a caller may match a substring
// against. Same rule as SortColumns: the column name is interpolated into SQL
// rather than bound, so nothing outside this map may reach the query.
var FilterColumns = map[string]string{
	"name":    "d.name",
	"address": "d.address",
	"phone":   "d.phone",
}

// SortColumns maps the sort keys the API accepts onto the columns they order
// by. A column name can't be passed to Postgres as a bind parameter, so it is
// interpolated into the query — every value that reaches ORDER BY must come
// from this map and never straight from the request.
var SortColumns = map[string]string{
	"name":       "d.name",
	"address":    "d.address",
	"phone":      "d.phone",
	"is_active":  "d.is_active",
	"created_at": "d.created_at",
}

const DefaultSortKey = "name"

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]dormdomain.Dormitory, error)
	ListActive(ctx context.Context, requesterID uuid.UUID, search string, limit int) ([]dormdomain.Dormitory, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (dormdomain.Dormitory, error)
	CountRooms(ctx context.Context, id uuid.UUID) (int64, error)
	Create(ctx context.Context, input CreateInput) (dormdomain.Dormitory, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (dormdomain.Dormitory, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

// ActivityLogger records dormitory create/update/delete events for the audit
// trail. Failures to record are logged but never block the dormitory flow.
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
// never fail the dormitory CRUD flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, entityID uuid.UUID, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "dormitory",
		EntityID:    &entityID,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]dormdomain.Dormitory, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	dormitories, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return dormitories, total, nil
}

func (s *Service) ListActive(ctx context.Context, requesterID uuid.UUID, search string, limit int) ([]dormdomain.Dormitory, error) {
	return s.repo.ListActive(ctx, requesterID, strings.TrimSpace(search), limit)
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (dormdomain.Dormitory, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) CheckDeletion(ctx context.Context, id, requesterID uuid.UUID) (DeletionCheck, error) {
	if _, err := s.repo.GetByID(ctx, id, requesterID); err != nil {
		return DeletionCheck{}, err
	}

	roomCount, err := s.repo.CountRooms(ctx, id)
	if err != nil {
		return DeletionCheck{}, err
	}

	return DeletionCheck{CanDelete: roomCount == 0, RoomCount: roomCount}, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (dormdomain.Dormitory, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Address = strings.TrimSpace(input.Address)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Description = strings.TrimSpace(input.Description)
	if input.Name == "" {
		return dormdomain.Dormitory{}, dormdomain.ErrRequiredDormitoryData
	}

	dormitory, err := s.repo.Create(ctx, input)
	if err != nil {
		return dormdomain.Dormitory{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", dormitory.ID, fmt.Sprintf("Created dormitory: %s", dormitory.Name), ipAddress)
	return dormitory, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (dormdomain.Dormitory, error) {
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return dormdomain.Dormitory{}, dormdomain.ErrRequiredDormitoryData
		}
		input.Name = &name
	}
	if input.Address != nil {
		address := strings.TrimSpace(*input.Address)
		input.Address = &address
	}
	if input.Phone != nil {
		phone := strings.TrimSpace(*input.Phone)
		input.Phone = &phone
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		input.Description = &description
	}

	dormitory, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return dormdomain.Dormitory{}, err
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", dormitory.ID, fmt.Sprintf("Updated dormitory: %s", dormitory.Name), ipAddress)
	return dormitory, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	dormitory, _ := s.repo.GetByID(ctx, id, requesterID)

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", id, fmt.Sprintf("Deleted dormitory: %s", dormitory.Name), ipAddress)
	return nil
}
