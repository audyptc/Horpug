package usecase

import (
	"context"
	"strings"
	"time"

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

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
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

func (s *Service) Create(ctx context.Context, input CreateInput) (metdomain.Meter, error) {
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

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (metdomain.Meter, error) {
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

	return s.repo.Update(ctx, id, requesterID, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
