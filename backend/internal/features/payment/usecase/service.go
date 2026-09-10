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

	Search string
	// Columns narrows individual columns by substring, keyed by FilterColumns.
	// Search casts a wide OR across tenant/room/dormitory/reference; these are
	// ANDed on top, so the two answer different questions and compose.
	Columns       map[string]string
	PaymentMethod *paymentdomain.PaymentMethod
	SortKey       string
	SortDesc      bool
}

// FilterColumns whitelists the columns a caller may match a substring against,
// passed as f[<column>]=value. The repository resolves each key to the actual
// SQL it needs (a plain column, a concatenation, or an EXISTS subquery for the
// per-item reference number) — this map exists purely so an unrecognised
// column is rejected rather than silently ignored.
var FilterColumns = map[string]string{
	"tenant_name":    "tenant_name",
	"room_number":    "room_number",
	"dormitory_name": "dormitory_name",
	"reference_no":   "reference_no",
}

// SortColumns whitelists the sort keys the API accepts. As with FilterColumns,
// the repository resolves each key to its actual ORDER BY expression.
var SortColumns = map[string]string{
	"tenant_name":  "tenant_name",
	"room_number":  "room_number",
	"amount":       "amount",
	"payment_date": "payment_date",
}

const DefaultSortKey = "payment_date"

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
