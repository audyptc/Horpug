package usecase

import (
	"context"
	"strings"
	"time"

	parceldomain "apihorpug/internal/features/parcel/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	TenantID    *uuid.UUID
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
	Status      *parceldomain.ParcelStatus

	Search string
	// Columns narrows individual columns by substring, keyed by FilterColumns.
	// Search casts a wide OR across tenant, room, dormitory, courier and
	// tracking number; these are ANDed on top, so the two answer different
	// questions and compose.
	Columns  map[string]string
	SortKey  string
	SortDesc bool
}

// FilterColumns whitelists the columns a caller may match a substring against,
// passed as f[<column>]=value. The repository resolves each key to the actual
// SQL it needs — this map exists purely so an unrecognised column is rejected
// rather than silently ignored.
var FilterColumns = map[string]string{
	"tenant_name":     "tenant_name",
	"room_number":     "room_number",
	"courier":         "courier",
	"tracking_number": "tracking_number",
}

// SortColumns whitelists the sort keys the API accepts. As with FilterColumns,
// the repository resolves each key to its actual ORDER BY expression.
var SortColumns = map[string]string{
	"tenant_name":     "tenant_name",
	"room_number":     "room_number",
	"courier":         "courier",
	"tracking_number": "tracking_number",
	"status":          "status",
	"received_date":   "received_date",
}

const DefaultSortKey = "received_date"

type CreateInput struct {
	TenantID       uuid.UUID
	RoomID         *uuid.UUID
	Courier        string
	TrackingNumber string
	Status         parceldomain.ParcelStatus
	ReceivedDate   time.Time
	Note           string
	CreatedBy      *uuid.UUID
}

type UpdateInput struct {
	RoomID         *uuid.UUID
	Courier        *string
	TrackingNumber *string
	Status         *parceldomain.ParcelStatus
	ReceivedDate   *time.Time
	Note           *string
	UpdatedBy      *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]parceldomain.Parcel, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (parceldomain.Parcel, error)
	Create(ctx context.Context, input CreateInput) (parceldomain.Parcel, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (parceldomain.Parcel, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]parceldomain.Parcel, int64, error) {
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

	parcels, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return parcels, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (parceldomain.Parcel, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (parceldomain.Parcel, error) {
	input.Courier = strings.TrimSpace(input.Courier)
	input.TrackingNumber = strings.TrimSpace(input.TrackingNumber)
	input.Note = strings.TrimSpace(input.Note)
	if input.Status == "" {
		input.Status = parceldomain.ParcelStatusPending
	}
	if input.RoomID != nil && *input.RoomID == uuid.Nil {
		input.RoomID = nil
	}

	if input.TenantID == uuid.Nil || input.ReceivedDate.IsZero() {
		return parceldomain.Parcel{}, parceldomain.ErrRequiredParcelData
	}
	if !input.Status.Valid() {
		return parceldomain.Parcel{}, parceldomain.ErrInvalidParcelStatus
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (parceldomain.Parcel, error) {
	if input.Status != nil && !input.Status.Valid() {
		return parceldomain.Parcel{}, parceldomain.ErrInvalidParcelStatus
	}
	if input.Courier != nil {
		courier := strings.TrimSpace(*input.Courier)
		input.Courier = &courier
	}
	if input.TrackingNumber != nil {
		trackingNumber := strings.TrimSpace(*input.TrackingNumber)
		input.TrackingNumber = &trackingNumber
	}
	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		input.Note = &note
	}

	return s.repo.Update(ctx, id, requesterID, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
