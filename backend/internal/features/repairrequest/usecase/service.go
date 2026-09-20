package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	repairrequestdomain "apihorpug/internal/features/repairrequest/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
	TenantID    *uuid.UUID
	Category    *repairrequestdomain.RepairCategory
	Status      *repairrequestdomain.RepairStatus

	Search string
	// Columns narrows individual columns by substring, keyed by FilterColumns.
	// Search casts a wide OR across room, dormitory, tenant and description;
	// these are ANDed on top, so the two answer different questions and
	// compose.
	Columns  map[string]string
	SortKey  string
	SortDesc bool
}

// FilterColumns whitelists the columns a caller may match a substring against,
// passed as f[<column>]=value. The repository resolves each key to the actual
// SQL it needs — this map exists purely so an unrecognised column is rejected
// rather than silently ignored.
var FilterColumns = map[string]string{
	"room_number":    "room_number",
	"dormitory_name": "dormitory_name",
	"tenant_name":    "tenant_name",
	"description":    "description",
}

// SortColumns whitelists the sort keys the API accepts. As with FilterColumns,
// the repository resolves each key to its actual ORDER BY expression.
var SortColumns = map[string]string{
	"room_number":   "room_number",
	"tenant_name":   "tenant_name",
	"category":      "category",
	"status":        "status",
	"reported_date": "reported_date",
	"description":   "description",
}

const DefaultSortKey = "reported_date"

type CreateInput struct {
	RoomID       uuid.UUID
	TenantID     *uuid.UUID
	Category     repairrequestdomain.RepairCategory
	Description  string
	Status       repairrequestdomain.RepairStatus
	ReportedDate time.Time
	CreatedBy    *uuid.UUID
}

type UpdateInput struct {
	TenantID *uuid.UUID
	// ClearTenant detaches the reporting tenant. A nil TenantID only means
	// "leave unchanged", so removing the tenant needs its own signal.
	ClearTenant  bool
	Category     *repairrequestdomain.RepairCategory
	Description  *string
	Status       *repairrequestdomain.RepairStatus
	ReportedDate *time.Time
	UpdatedBy    *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]repairrequestdomain.RepairRequest, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (repairrequestdomain.RepairRequest, error)
	Create(ctx context.Context, input CreateInput) (repairrequestdomain.RepairRequest, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (repairrequestdomain.RepairRequest, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

// ActivityLogger records who created, changed or deleted a repair request for
// the audit trail. Failures to record are logged but never block the flow.
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
// never fail the repair request flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, request repairrequestdomain.RepairRequest, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	var dormitoryRef *uuid.UUID
	if request.DormitoryID != uuid.Nil {
		dormitoryRef = &request.DormitoryID
	}
	entityID := request.ID
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "repair_request",
		EntityID:    &entityID,
		DormitoryID: dormitoryRef,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func repairActivityDescription(request repairrequestdomain.RepairRequest) string {
	return fmt.Sprintf("repair request: room %s, %s, %s", request.RoomNumber, request.Category, request.Status)
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]repairrequestdomain.RepairRequest, int64, error) {
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

	requests, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (repairrequestdomain.RepairRequest, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (repairrequestdomain.RepairRequest, error) {
	input.Description = strings.TrimSpace(input.Description)
	if input.Category == "" {
		input.Category = repairrequestdomain.RepairCategoryOther
	}
	if input.Status == "" {
		input.Status = repairrequestdomain.RepairStatusPending
	}

	if input.RoomID == uuid.Nil || input.Description == "" || input.ReportedDate.IsZero() {
		return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrRequiredRepairRequestData
	}
	if !input.Category.Valid() {
		return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrInvalidCategory
	}
	if !input.Status.Valid() {
		return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrInvalidStatus
	}

	request, err := s.repo.Create(ctx, input)
	if err != nil {
		return repairrequestdomain.RepairRequest{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", request, "Created "+repairActivityDescription(request), ipAddress)
	return request, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (repairrequestdomain.RepairRequest, error) {
	if input.Category != nil && !input.Category.Valid() {
		return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrInvalidCategory
	}
	if input.Status != nil && !input.Status.Valid() {
		return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrInvalidStatus
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		if description == "" {
			return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrRequiredRepairRequestData
		}
		input.Description = &description
	}

	// Read before the update so the log can show what the status changed from.
	// It also checks the requester's access, so a missing or out-of-scope
	// request fails here as not found, same as the update itself would.
	before, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return repairrequestdomain.RepairRequest{}, err
	}

	request, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return repairrequestdomain.RepairRequest{}, err
	}

	description := "Updated " + repairActivityDescription(request)
	if before.Status != request.Status {
		description = fmt.Sprintf("Updated repair request: room %s, %s, %s -> %s", request.RoomNumber, request.Category, before.Status, request.Status)
	}
	s.recordActivity(ctx, &requesterID, "UPDATE", request, description, ipAddress)
	return request, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	// Read first: once deleted there is no room or category left to describe.
	request, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", request, "Deleted "+repairActivityDescription(request), ipAddress)
	return nil
}
