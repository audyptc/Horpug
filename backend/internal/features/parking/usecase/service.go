package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	parkingdomain "apihorpug/internal/features/parking/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	TenantID    *uuid.UUID
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
	VehicleType *parkingdomain.VehicleType

	Search string
	// Columns narrows individual columns by substring, keyed by FilterColumns.
	// Search casts a wide OR across tenant, room, dormitory, license plate and
	// parking spot; these are ANDed on top, so the two answer different
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
	"tenant_name":   "tenant_name",
	"room_number":   "room_number",
	"license_plate": "license_plate",
	"parking_spot":  "parking_spot",
}

// SortColumns whitelists the sort keys the API accepts. As with FilterColumns,
// the repository resolves each key to its actual ORDER BY expression.
var SortColumns = map[string]string{
	"tenant_name":   "tenant_name",
	"room_number":   "room_number",
	"vehicle_type":  "vehicle_type",
	"license_plate": "license_plate",
	"parking_spot":  "parking_spot",
	"created_at":    "created_at",
}

const DefaultSortKey = "created_at"

type CreateInput struct {
	TenantID     uuid.UUID
	RoomID       *uuid.UUID
	VehicleType  parkingdomain.VehicleType
	LicensePlate string
	ParkingSpot  string
	CreatedBy    *uuid.UUID
}

type UpdateInput struct {
	RoomID       *uuid.UUID
	VehicleType  *parkingdomain.VehicleType
	LicensePlate *string
	ParkingSpot  *string
	UpdatedBy    *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]parkingdomain.Parking, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (parkingdomain.Parking, error)
	Create(ctx context.Context, input CreateInput) (parkingdomain.Parking, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (parkingdomain.Parking, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

// ActivityLogger records who registered, changed or deleted a parking
// registration for the audit trail. Failures to record are logged but never
// block the flow.
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
// never fail the parking flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, parking parkingdomain.Parking, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	entityID := parking.ID
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "parking",
		EntityID:    &entityID,
		DormitoryID: parking.DormitoryID,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func parkingActivityDescription(parking parkingdomain.Parking) string {
	return fmt.Sprintf("parking: %s, %s, %s", parking.TenantName, parking.VehicleType, parking.LicensePlate)
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]parkingdomain.Parking, int64, error) {
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

	parkings, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return parkings, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (parkingdomain.Parking, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (parkingdomain.Parking, error) {
	input.LicensePlate = strings.TrimSpace(input.LicensePlate)
	input.ParkingSpot = strings.TrimSpace(input.ParkingSpot)
	if input.VehicleType == "" {
		input.VehicleType = parkingdomain.VehicleTypeMotorcycle
	}
	if input.RoomID != nil && *input.RoomID == uuid.Nil {
		input.RoomID = nil
	}

	if input.TenantID == uuid.Nil || input.LicensePlate == "" {
		return parkingdomain.Parking{}, parkingdomain.ErrRequiredParkingData
	}
	if !input.VehicleType.Valid() {
		return parkingdomain.Parking{}, parkingdomain.ErrInvalidVehicleType
	}

	parking, err := s.repo.Create(ctx, input)
	if err != nil {
		return parkingdomain.Parking{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", parking, "Created "+parkingActivityDescription(parking), ipAddress)
	return parking, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (parkingdomain.Parking, error) {
	if input.VehicleType != nil && !input.VehicleType.Valid() {
		return parkingdomain.Parking{}, parkingdomain.ErrInvalidVehicleType
	}
	if input.LicensePlate != nil {
		licensePlate := strings.TrimSpace(*input.LicensePlate)
		if licensePlate == "" {
			return parkingdomain.Parking{}, parkingdomain.ErrRequiredParkingData
		}
		input.LicensePlate = &licensePlate
	}
	if input.ParkingSpot != nil {
		parkingSpot := strings.TrimSpace(*input.ParkingSpot)
		input.ParkingSpot = &parkingSpot
	}

	parking, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return parkingdomain.Parking{}, err
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", parking, "Updated "+parkingActivityDescription(parking), ipAddress)
	return parking, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	// Read first: once deleted there is no tenant or plate left to describe.
	// It also checks the requester's access, so a missing or out-of-scope
	// registration fails here as not found, same as the delete itself would.
	parking, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", parking, "Deleted "+parkingActivityDescription(parking), ipAddress)
	return nil
}
