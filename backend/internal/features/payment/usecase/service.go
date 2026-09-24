package usecase

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
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

// UpdateInput replaces a payment's date, note and items wholesale. The invoice
// a payment belongs to is deliberately not editable: moving a payment to a
// different invoice is a delete and a fresh Create.
type UpdateInput struct {
	PaymentDate time.Time
	Note        string
	Items       []ItemInput
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]paymentdomain.Payment, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (paymentdomain.Payment, error)
	// Create, Update and Delete also report the invoice status flip the change
	// caused (if any), so it can be written into the activity log.
	Create(ctx context.Context, input CreateInput) (paymentdomain.Payment, paymentdomain.InvoiceStatusChange, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (paymentdomain.Payment, paymentdomain.InvoiceStatusChange, error)
	Void(ctx context.Context, id, requesterID uuid.UUID, reason string) (paymentdomain.InvoiceStatusChange, error)
	ReceiptRepository
}

// ActivityLogger records who created, changed or deleted a payment for the
// audit trail. Failures to record are logged but never block the payment flow.
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
// never fail the payment flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, payment paymentdomain.Payment, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	var dormitoryRef *uuid.UUID
	if payment.DormitoryID != uuid.Nil {
		dormitoryRef = &payment.DormitoryID
	}
	entityID := payment.ID
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "payment",
		EntityID:    &entityID,
		DormitoryID: dormitoryRef,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func paymentActivityDescription(payment paymentdomain.Payment) string {
	return fmt.Sprintf("payment %s: room %s, %.2f THB", payment.ReceiptNo, payment.RoomNumber, payment.TotalAmount)
}

// withInvoiceStatusChange appends the invoice status flip, when there was one.
func withInvoiceStatusChange(description string, change paymentdomain.InvoiceStatusChange) string {
	if !change.Changed() {
		return description
	}
	return fmt.Sprintf("%s — invoice status: %s -> %s", description, change.From, change.To)
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

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (paymentdomain.Payment, error) {
	input.Note = strings.TrimSpace(input.Note)

	if input.InvoiceID == uuid.Nil || input.PaymentDate.IsZero() {
		return paymentdomain.Payment{}, paymentdomain.ErrRequiredPaymentData
	}
	if err := normalizeItems(input.Items); err != nil {
		return paymentdomain.Payment{}, err
	}

	payment, change, err := s.repo.Create(ctx, input)
	if err != nil {
		return paymentdomain.Payment{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", payment,
		withInvoiceStatusChange("Created "+paymentActivityDescription(payment), change), ipAddress)
	return payment, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (paymentdomain.Payment, error) {
	input.Note = strings.TrimSpace(input.Note)

	if input.PaymentDate.IsZero() {
		return paymentdomain.Payment{}, paymentdomain.ErrRequiredPaymentData
	}
	if err := normalizeItems(input.Items); err != nil {
		return paymentdomain.Payment{}, err
	}

	// Read before the update so the log can show what the amount changed from.
	// It also checks the requester's access, so a missing or out-of-scope
	// payment fails here as not found, same as the update itself would.
	before, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return paymentdomain.Payment{}, err
	}
	if before.Status == paymentdomain.PaymentStatusVoided {
		return paymentdomain.Payment{}, paymentdomain.ErrPaymentVoided
	}
	if changesReceipt(before, input) {
		return paymentdomain.Payment{}, paymentdomain.ErrReceiptIssued
	}

	payment, change, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return paymentdomain.Payment{}, err
	}

	description := fmt.Sprintf("Updated payment: room %s, %.2f -> %.2f THB", payment.RoomNumber, before.TotalAmount, payment.TotalAmount)
	s.recordActivity(ctx, &requesterID, "UPDATE", payment, withInvoiceStatusChange(description, change), ipAddress)
	return payment, nil
}

// normalizeItems validates a payment's method lines and tidies them in place
// (trimmed reference, cash as the default method).
func normalizeItems(items []ItemInput) error {
	if len(items) == 0 {
		return paymentdomain.ErrRequiredItems
	}

	for i, item := range items {
		item.ReferenceNo = strings.TrimSpace(item.ReferenceNo)
		if item.PaymentMethod == "" {
			item.PaymentMethod = paymentdomain.PaymentMethodCash
		}
		if item.Amount <= 0 {
			return paymentdomain.ErrInvalidAmount
		}
		if !item.PaymentMethod.Valid() {
			return paymentdomain.ErrInvalidMethod
		}
		items[i] = item
	}

	return nil
}

// changesReceipt reports whether an edit would alter what the payment's issued
// receipt states: its date, or the method and amount of any line. The note
// and reference numbers are bookkeeping and may still be corrected.
func changesReceipt(before paymentdomain.Payment, input UpdateInput) bool {
	if !sameDay(before.PaymentDate, input.PaymentDate) || len(before.Items) != len(input.Items) {
		return true
	}
	for i, item := range input.Items {
		old := before.Items[i]
		if old.PaymentMethod != item.PaymentMethod || math.Abs(old.Amount-item.Amount) >= 0.005 {
			return true
		}
	}
	return false
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.UTC().Date()
	by, bm, bd := b.UTC().Date()
	return ay == by && am == bm && ad == bd
}

// Void cancels a payment's receipt. The payment stays on record with the
// reason, stops counting towards its invoice, and its number is never reused.
func (s *Service) Void(ctx context.Context, id, requesterID uuid.UUID, reason, ipAddress string) error {
	reason = strings.TrimSpace(reason)

	// Read first so the log can describe the payment being voided.
	payment, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return err
	}
	if payment.Status == paymentdomain.PaymentStatusVoided {
		return paymentdomain.ErrPaymentVoided
	}

	change, err := s.repo.Void(ctx, id, requesterID, reason)
	if err != nil {
		return err
	}

	description := "Voided " + paymentActivityDescription(payment)
	if reason != "" {
		description += " (reason: " + reason + ")"
	}
	s.recordActivity(ctx, &requesterID, "VOID", payment, withInvoiceStatusChange(description, change), ipAddress)
	return nil
}
