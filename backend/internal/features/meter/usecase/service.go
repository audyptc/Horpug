package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	meterdomain "apihorpug/internal/features/meter/domain"

	"github.com/google/uuid"
)

// ListFilters narrows and orders a meter listing. The nil-able fields mean "no
// preference" rather than a zero value, so asking for unbilled readings stays
// distinct from not filtering on billing at all.
type ListFilters struct {
	RoomID        *uuid.UUID
	DormitoryID   *uuid.UUID
	BillingMethod *meterdomain.BillingMethod
	IsBilled      *bool
	Search        string
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
	"room_number":    "rm.room_number",
	"dormitory_name": "d.name",
}

// IsBilledExpr reports whether a reading has already been charged on an
// invoice. It's derived rather than stored, so filtering and sorting on it
// have to repeat the subquery the listing selects.
const IsBilledExpr = `EXISTS (SELECT 1 FROM invoice_items ii WHERE ii.reference_id = em.id AND ii.item_type = 'electricity')`

// SortColumns maps the sort keys the API accepts onto the expressions they
// order by. A column name can't be passed to Postgres as a bind parameter, so
// it is interpolated into the query — every value that reaches ORDER BY must
// come from this map and never straight from the request.
var SortColumns = map[string]string{
	"room_number":    "rm.room_number",
	"dormitory_name": "d.name",
	"reading_date":   "em.reading_date",
	"billing_method": "em.billing_method",
	"previous_unit":  "em.previous_unit",
	"current_unit":   "em.current_unit",
	"unit_used":      "em.unit_used",
	"price_per_unit": "em.price_per_unit",
	"total_amount":   "em.total_amount",
	"is_billed":      IsBilledExpr,
	"created_at":     "em.created_at",
}

// Newest reading first is the useful default here: a meter listing is read as
// a running log, not as an alphabetical register.
const DefaultSortKey = "reading_date"

type CreateInput struct {
	RoomID        uuid.UUID
	BillingMethod meterdomain.BillingMethod
	ReadingDate   time.Time
	PreviousUnit  float64
	CurrentUnit   float64
	PricePerUnit  float64
	FlatAmount    *float64
	Note          string
	CreatedBy     *uuid.UUID
}

type UpdateInput struct {
	BillingMethod *meterdomain.BillingMethod
	ReadingDate   *time.Time
	PreviousUnit  *float64
	CurrentUnit   *float64
	PricePerUnit  *float64
	FlatAmount    *float64
	Note          *string
	UpdatedBy     *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]meterdomain.Meter, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (meterdomain.Meter, error)
	Create(ctx context.Context, input CreateInput) (meterdomain.Meter, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (meterdomain.Meter, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

// ActivityLogger records meter reading create/update/delete events for the
// audit trail. Failures to record are logged but never block the meter flow.
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
// never fail the meter CRUD flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, entityID uuid.UUID, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "meter",
		EntityID:    &entityID,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func meterActivityDescription(meter meterdomain.Meter) string {
	return fmt.Sprintf("electricity meter reading: room %s (%s)", meter.RoomNumber, meter.ReadingDate.Format("2006-01-02"))
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]meterdomain.Meter, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	meters, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return meters, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (meterdomain.Meter, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (meterdomain.Meter, error) {
	input.Note = strings.TrimSpace(input.Note)

	if input.RoomID == uuid.Nil || input.ReadingDate.IsZero() {
		return meterdomain.Meter{}, meterdomain.ErrRequiredMeterData
	}
	if input.BillingMethod == "" {
		input.BillingMethod = meterdomain.BillingMethodMetered
	}
	if !input.BillingMethod.Valid() {
		return meterdomain.Meter{}, meterdomain.ErrInvalidBillingMethod
	}
	if input.PreviousUnit < 0 || input.CurrentUnit < 0 || input.CurrentUnit < input.PreviousUnit {
		return meterdomain.Meter{}, meterdomain.ErrInvalidMeterUnits
	}
	if input.PricePerUnit < 0 {
		return meterdomain.Meter{}, meterdomain.ErrInvalidMeterPrice
	}

	if input.BillingMethod == meterdomain.BillingMethodFlat {
		if input.FlatAmount == nil || *input.FlatAmount < 0 {
			return meterdomain.Meter{}, meterdomain.ErrRequiredFlatAmount
		}
	} else {
		// total_amount for metered readings is always the units×price
		// formula; ignore any flat_amount supplied alongside it.
		input.FlatAmount = nil
	}

	meter, err := s.repo.Create(ctx, input)
	if err != nil {
		return meterdomain.Meter{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", meter.ID, "Created "+meterActivityDescription(meter), ipAddress)
	return meter, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (meterdomain.Meter, error) {
	if input.BillingMethod != nil && !input.BillingMethod.Valid() {
		return meterdomain.Meter{}, meterdomain.ErrInvalidBillingMethod
	}
	if input.PreviousUnit != nil && *input.PreviousUnit < 0 {
		return meterdomain.Meter{}, meterdomain.ErrInvalidMeterUnits
	}
	if input.CurrentUnit != nil && *input.CurrentUnit < 0 {
		return meterdomain.Meter{}, meterdomain.ErrInvalidMeterUnits
	}
	if input.PreviousUnit != nil && input.CurrentUnit != nil && *input.CurrentUnit < *input.PreviousUnit {
		return meterdomain.Meter{}, meterdomain.ErrInvalidMeterUnits
	}
	if input.PricePerUnit != nil && *input.PricePerUnit < 0 {
		return meterdomain.Meter{}, meterdomain.ErrInvalidMeterPrice
	}
	if input.FlatAmount != nil && *input.FlatAmount < 0 {
		return meterdomain.Meter{}, meterdomain.ErrRequiredFlatAmount
	}
	// Switching a reading to flat billing requires flat_amount in the same
	// request: without it there's no prior value to fall back to that the
	// usecase layer can see.
	if input.BillingMethod != nil && *input.BillingMethod == meterdomain.BillingMethodFlat && input.FlatAmount == nil {
		return meterdomain.Meter{}, meterdomain.ErrRequiredFlatAmount
	}
	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		input.Note = &note
	}

	meter, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return meterdomain.Meter{}, err
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", meter.ID, "Updated "+meterActivityDescription(meter), ipAddress)
	return meter, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	meter, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", id, "Deleted "+meterActivityDescription(meter), ipAddress)
	return nil
}
