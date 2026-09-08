package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	contractdomain "apihorpug/internal/features/contract/domain"
	roomdomain "apihorpug/internal/features/room/domain"

	"github.com/google/uuid"
)

// ListFilters narrows and orders a contract listing. The nil-able fields mean
// "no preference" rather than a zero value, so asking for terminated contracts
// stays distinct from not filtering on status at all.
type ListFilters struct {
	TenantID    *uuid.UUID
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
	Status      *contractdomain.ContractStatus
	Search      string
	// Columns narrows individual columns by substring, keyed by FilterColumns.
	// Search casts a wide OR across every text column; these are ANDed on top,
	// so the two answer different questions and compose.
	Columns  map[string]string
	SortKey  string
	SortDesc bool
}

// FilterColumns whitelists the columns a caller may match a substring against.
// Same rule as SortColumns: the expression is interpolated into SQL rather
// than bound, so nothing outside this map may reach the query.
var FilterColumns = map[string]string{
	"tenant_name":    "(t.first_name || ' ' || t.last_name)",
	"room_number":    "rm.room_number",
	"dormitory_name": "d.name",
}

// SortColumns maps the sort keys the API accepts onto the expressions they
// order by. A column name can't be passed to Postgres as a bind parameter, so
// it is interpolated into the query — every value that reaches ORDER BY must
// come from this map and never straight from the request.
var SortColumns = map[string]string{
	"tenant_name":    "(t.first_name || ' ' || t.last_name)",
	"room_number":    "rm.room_number",
	"dormitory_name": "d.name",
	"start_date":     "c.start_date",
	"end_date":       "c.end_date",
	"rent_price":     "c.rent_price",
	"deposit":        "c.deposit",
	"status":         "c.status",
	"created_at":     "c.created_at",
}

const DefaultSortKey = "created_at"

type CreateInput struct {
	TenantID     uuid.UUID
	RoomID       uuid.UUID
	StartDate    time.Time
	EndDate      *time.Time
	RentPrice    float64
	Deposit      float64
	NumOccupants int
	Note         string
	CreatedBy    *uuid.UUID
}

type UpdateInput struct {
	StartDate    *time.Time
	EndDate      *time.Time
	RentPrice    *float64
	Deposit      *float64
	NumOccupants *int
	Status       *contractdomain.ContractStatus
	Note         *string
	UpdatedBy    *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]contractdomain.Contract, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (contractdomain.Contract, error)
	Create(ctx context.Context, input CreateInput) (contractdomain.Contract, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (contractdomain.Contract, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

// ActivityLogger records contract create/update/delete events for the audit
// trail. Failures to record are logged but never block the contract flow.
type ActivityLogger interface {
	Create(ctx context.Context, input activitylogusecase.CreateInput) (activitylogdomain.ActivityLog, error)
}

// RoomStatusUpdater keeps a room's status in step with its contract
// lifecycle (occupied while a contract is active, available once it's
// expired/terminated). Failures are logged but never block the contract
// flow — rm.status is a convenience display field, not the source of truth
// for occupancy (that's still the contracts table itself).
type RoomStatusUpdater interface {
	SetStatus(ctx context.Context, roomID uuid.UUID, status roomdomain.RoomStatus) error
}

type Service struct {
	repo        Repository
	activityLog ActivityLogger
	roomStatus  RoomStatusUpdater
}

func New(repo Repository, activityLog ActivityLogger, roomStatus RoomStatusUpdater) *Service {
	return &Service{repo: repo, activityLog: activityLog, roomStatus: roomStatus}
}

// syncRoomStatus is best-effort: a failure to update the room's display
// status must never fail the contract CRUD flow itself.
func (s *Service) syncRoomStatus(ctx context.Context, roomID uuid.UUID, status roomdomain.RoomStatus) {
	if s.roomStatus == nil {
		return
	}
	if err := s.roomStatus.SetStatus(ctx, roomID, status); err != nil {
		log.Printf("failed to sync room status (room=%s, status=%s): %v", roomID, status, err)
	}
}

// recordActivity is best-effort: a failure to write the audit trail must
// never fail the contract CRUD flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, entityID uuid.UUID, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "contract",
		EntityID:    &entityID,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]contractdomain.Contract, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	contracts, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return contracts, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (contractdomain.Contract, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (contractdomain.Contract, error) {
	input.Note = strings.TrimSpace(input.Note)

	if input.TenantID == uuid.Nil || input.RoomID == uuid.Nil || input.StartDate.IsZero() {
		return contractdomain.Contract{}, contractdomain.ErrRequiredContractData
	}
	if input.EndDate != nil && input.EndDate.Before(input.StartDate) {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractDates
	}
	if input.RentPrice < 0 || input.Deposit < 0 {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractAmount
	}
	if input.NumOccupants <= 0 {
		input.NumOccupants = 1
	}

	contract, err := s.repo.Create(ctx, input)
	if err != nil {
		return contractdomain.Contract{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", contract.ID,
		fmt.Sprintf("Created contract: %s - room %s", contract.TenantName, contract.RoomNumber), ipAddress)
	s.syncRoomStatus(ctx, contract.RoomID, roomdomain.RoomStatusOccupied)
	return contract, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (contractdomain.Contract, error) {
	if input.StartDate != nil && input.StartDate.IsZero() {
		return contractdomain.Contract{}, contractdomain.ErrRequiredContractData
	}
	if input.StartDate != nil && input.EndDate != nil && input.EndDate.Before(*input.StartDate) {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractDates
	}
	if input.RentPrice != nil && *input.RentPrice < 0 {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractAmount
	}
	if input.Deposit != nil && *input.Deposit < 0 {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractAmount
	}
	if input.NumOccupants != nil && *input.NumOccupants <= 0 {
		return contractdomain.Contract{}, contractdomain.ErrInvalidNumOccupants
	}
	if input.Status != nil && !input.Status.Valid() {
		return contractdomain.Contract{}, contractdomain.ErrInvalidContractStatus
	}
	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		input.Note = &note
	}

	contract, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return contractdomain.Contract{}, err
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", contract.ID,
		fmt.Sprintf("Updated contract: %s - room %s", contract.TenantName, contract.RoomNumber), ipAddress)
	if input.Status != nil {
		if *input.Status == contractdomain.ContractStatusActive {
			s.syncRoomStatus(ctx, contract.RoomID, roomdomain.RoomStatusOccupied)
		} else {
			s.syncRoomStatus(ctx, contract.RoomID, roomdomain.RoomStatusAvailable)
		}
	}
	return contract, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	contract, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return err
	}
	if contract.Status == contractdomain.ContractStatusActive {
		return contractdomain.ErrContractIsActive
	}

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", id,
		fmt.Sprintf("Deleted contract: %s - room %s", contract.TenantName, contract.RoomNumber), ipAddress)
	return nil
}
