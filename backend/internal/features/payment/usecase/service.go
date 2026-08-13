package usecase

import (
	"context"
	"strings"
	"time"

	paymentdomain "apihorpug/internal/features/payment/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	InvoiceID   *uuid.UUID
	ContractID  *uuid.UUID
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
	TenantID    *uuid.UUID
	DateFrom    *time.Time
	DateTo      *time.Time
}

// ItemInput is one payment-method line to record as part of a Create call,
// e.g. the "cash 3,000" portion of a receipt split across multiple methods.
type ItemInput struct {
	PaymentMethod paymentdomain.PaymentMethod
	Amount        float64
	ReferenceNo   string
}

type CreateInput struct {
	InvoiceID   uuid.UUID
	PaymentDate time.Time
	Note        string
	Items       []ItemInput
	CreatedBy   *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]paymentdomain.Payment, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (paymentdomain.Payment, error)
	Create(ctx context.Context, input CreateInput) (paymentdomain.Payment, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]paymentdomain.Payment, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	payments, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (paymentdomain.Payment, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (paymentdomain.Payment, error) {
	input.Note = strings.TrimSpace(input.Note)

	if input.InvoiceID == uuid.Nil || input.PaymentDate.IsZero() {
		return paymentdomain.Payment{}, paymentdomain.ErrRequiredPaymentData
	}
	if len(input.Items) == 0 {
		return paymentdomain.Payment{}, paymentdomain.ErrRequiredItems
	}

	for i, item := range input.Items {
		item.ReferenceNo = strings.TrimSpace(item.ReferenceNo)
		if item.PaymentMethod == "" {
			item.PaymentMethod = paymentdomain.PaymentMethodCash
		}
		if item.Amount <= 0 {
			return paymentdomain.Payment{}, paymentdomain.ErrInvalidAmount
		}
		if !item.PaymentMethod.Valid() {
			return paymentdomain.Payment{}, paymentdomain.ErrInvalidMethod
		}
		input.Items[i] = item
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
