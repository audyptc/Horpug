package usecase

import (
	"context"
	"strings"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	ContractID  *uuid.UUID
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
	TenantID    *uuid.UUID
	Status      *invoicedomain.InvoiceStatus
	PeriodYear  *int
	PeriodMonth *int
}

type CreateInput struct {
	ContractID  uuid.UUID
	PeriodYear  int
	PeriodMonth int
	IssueDate   time.Time
	DueDate     time.Time
	Note        string
	CreatedBy   *uuid.UUID
}

type UpdateInput struct {
	Status    *invoicedomain.InvoiceStatus
	DueDate   *time.Time
	Note      *string
	UpdatedBy *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]invoicedomain.Invoice, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (invoicedomain.Invoice, error)
	Create(ctx context.Context, input CreateInput) (invoicedomain.Invoice, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (invoicedomain.Invoice, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]invoicedomain.Invoice, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	invoices, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (invoicedomain.Invoice, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (invoicedomain.Invoice, error) {
	input.Note = strings.TrimSpace(input.Note)

	if input.ContractID == uuid.Nil || input.PeriodYear <= 0 || input.IssueDate.IsZero() || input.DueDate.IsZero() {
		return invoicedomain.Invoice{}, invoicedomain.ErrRequiredInvoiceData
	}
	if input.PeriodMonth < 1 || input.PeriodMonth > 12 {
		return invoicedomain.Invoice{}, invoicedomain.ErrInvalidInvoicePeriod
	}
	if input.DueDate.Before(input.IssueDate) {
		return invoicedomain.Invoice{}, invoicedomain.ErrInvalidInvoiceDates
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (invoicedomain.Invoice, error) {
	if input.Status != nil && !input.Status.Valid() {
		return invoicedomain.Invoice{}, invoicedomain.ErrInvalidInvoiceStatus
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
