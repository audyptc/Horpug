package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
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

// ActivityLogger records who received, changed or deleted a parcel for the
// audit trail. Failures to record are logged but never block the flow.
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
// never fail the parcel flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, parcel parceldomain.Parcel, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	entityID := parcel.ID
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "parcel",
		EntityID:    &entityID,
		DormitoryID: parcel.DormitoryID,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func parcelActivityDescription(parcel parceldomain.Parcel) string {
	return fmt.Sprintf("parcel: %s, %s, %s", parcel.TenantName, parcel.Courier, parcel.TrackingNumber)
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

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (parceldomain.Parcel, error) {
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

	parcel, err := s.repo.Create(ctx, input)
	if err != nil {
		return parceldomain.Parcel{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", parcel, "Created "+parcelActivityDescription(parcel), ipAddress)
	return parcel, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (parceldomain.Parcel, error) {
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

	parcel, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return parceldomain.Parcel{}, err
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", parcel, "Updated "+parcelActivityDescription(parcel), ipAddress)
	return parcel, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	// Read first: once deleted there is no tenant or tracking number left to
	// describe. It also checks the requester's access, so a missing or
	// out-of-scope parcel fails here as not found, same as the delete would.
	parcel, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", parcel, "Deleted "+parcelActivityDescription(parcel), ipAddress)
	return nil
}
