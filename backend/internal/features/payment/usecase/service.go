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

type CreateInput struct {
	InvoiceID     uuid.UUID
	Amount        float64
	PaymentMethod paymentdomain.PaymentMethod
	PaymentDate   time.Time
	ReferenceNo   string
	Note          string
	CreatedBy     *uuid.UUID
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
	input.ReferenceNo = strings.TrimSpace(input.ReferenceNo)
	input.Note = strings.TrimSpace(input.Note)
	if input.PaymentMethod == "" {
		input.PaymentMethod = paymentdomain.PaymentMethodCash
	}

	if input.InvoiceID == uuid.Nil || input.PaymentDate.IsZero() {
		return paymentdomain.Payment{}, paymentdomain.ErrRequiredPaymentData
	}
	if input.Amount <= 0 {
		return paymentdomain.Payment{}, paymentdomain.ErrInvalidAmount
	}
	if !input.PaymentMethod.Valid() {
		return paymentdomain.Payment{}, paymentdomain.ErrInvalidMethod
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
