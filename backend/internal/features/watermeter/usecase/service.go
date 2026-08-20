package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	metdomain "apihorpug/internal/features/watermeter/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
}

type CreateInput struct {
	RoomID        uuid.UUID
	BillingMethod metdomain.BillingMethod
	ReadingDate   time.Time
	PreviousUnit  float64
	CurrentUnit   float64
	PricePerUnit  float64
	FlatAmount    *float64
	Note          string
	CreatedBy     *uuid.UUID
}

type UpdateInput struct {
	BillingMethod *metdomain.BillingMethod
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
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]metdomain.Meter, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (metdomain.Meter, error)
	Create(ctx context.Context, input CreateInput) (metdomain.Meter, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (metdomain.Meter, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

// ActivityLogger records water meter reading create/update/delete events for
// the audit trail. Failures to record are logged but never block the meter
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
// never fail the meter CRUD flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, entityID uuid.UUID, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "water_meter",
		EntityID:    &entityID,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func meterActivityDescription(meter metdomain.Meter) string {
	return fmt.Sprintf("water meter reading: room %s (%s)", meter.RoomNumber, meter.ReadingDate.Format("2006-01-02"))
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]metdomain.Meter, int64, error) {
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

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (metdomain.Meter, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (metdomain.Meter, error) {
	input.Note = strings.TrimSpace(input.Note)

	if input.RoomID == uuid.Nil || input.ReadingDate.IsZero() {
		return metdomain.Meter{}, metdomain.ErrRequiredMeterData
	}
	if input.BillingMethod == "" {
		input.BillingMethod = metdomain.BillingMethodMetered
	}
	if !input.BillingMethod.Valid() {
		return metdomain.Meter{}, metdomain.ErrInvalidBillingMethod
	}
	if input.PreviousUnit < 0 || input.CurrentUnit < 0 || input.CurrentUnit < input.PreviousUnit {
		return metdomain.Meter{}, metdomain.ErrInvalidMeterUnits
	}
	if input.PricePerUnit < 0 {
		return metdomain.Meter{}, metdomain.ErrInvalidMeterPrice
	}

	if input.BillingMethod == metdomain.BillingMethodFlat {
		if input.FlatAmount == nil || *input.FlatAmount < 0 {
			return metdomain.Meter{}, metdomain.ErrRequiredFlatAmount
		}
	} else {
		// total_amount for metered readings is always the units×price
		// formula; ignore any flat_amount supplied alongside it.
		input.FlatAmount = nil
	}

	meter, err := s.repo.Create(ctx, input)
	if err != nil {
		return metdomain.Meter{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", meter.ID, "Created "+meterActivityDescription(meter), ipAddress)
	return meter, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (metdomain.Meter, error) {
	if input.BillingMethod != nil && !input.BillingMethod.Valid() {
		return metdomain.Meter{}, metdomain.ErrInvalidBillingMethod
	}
	if input.PreviousUnit != nil && *input.PreviousUnit < 0 {
		return metdomain.Meter{}, metdomain.ErrInvalidMeterUnits
	}
	if input.CurrentUnit != nil && *input.CurrentUnit < 0 {
		return metdomain.Meter{}, metdomain.ErrInvalidMeterUnits
	}
	if input.PreviousUnit != nil && input.CurrentUnit != nil && *input.CurrentUnit < *input.PreviousUnit {
		return metdomain.Meter{}, metdomain.ErrInvalidMeterUnits
	}
	if input.PricePerUnit != nil && *input.PricePerUnit < 0 {
		return metdomain.Meter{}, metdomain.ErrInvalidMeterPrice
	}
	if input.FlatAmount != nil && *input.FlatAmount < 0 {
		return metdomain.Meter{}, metdomain.ErrRequiredFlatAmount
	}
	// Switching a reading to flat billing requires flat_amount in the same
	// request: without it there's no prior value to fall back to that the
	// usecase layer can see.
	if input.BillingMethod != nil && *input.BillingMethod == metdomain.BillingMethodFlat && input.FlatAmount == nil {
		return metdomain.Meter{}, metdomain.ErrRequiredFlatAmount
	}
	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		input.Note = &note
	}

	meter, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return metdomain.Meter{}, err
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
