package usecase

import (
	"context"
	"strings"
	"time"

	meterdomain "apihorpug/internal/features/meter/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
}

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

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
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

func (s *Service) Create(ctx context.Context, input CreateInput) (meterdomain.Meter, error) {
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

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (meterdomain.Meter, error) {
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

	return s.repo.Update(ctx, id, requesterID, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
