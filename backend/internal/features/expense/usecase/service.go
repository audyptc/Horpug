package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	expensedomain "apihorpug/internal/features/expense/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	DormitoryID *uuid.UUID
	Category    *expensedomain.ExpenseCategory
	DateFrom    *time.Time
	DateTo      *time.Time

	Search string
	// Columns narrows individual columns by substring, keyed by FilterColumns.
	// Search casts a wide OR across dormitory and description; these are ANDed
	// on top, so the two answer different questions and compose.
	Columns  map[string]string
	SortKey  string
	SortDesc bool
}

// FilterColumns whitelists the columns a caller may match a substring against,
// passed as f[<column>]=value. The repository resolves each key to the actual
// SQL it needs — this map exists purely so an unrecognised column is rejected
// rather than silently ignored.
var FilterColumns = map[string]string{
	"dormitory_name": "dormitory_name",
	"description":    "description",
}

// SortColumns whitelists the sort keys the API accepts. As with FilterColumns,
// the repository resolves each key to its actual ORDER BY expression.
var SortColumns = map[string]string{
	"dormitory_name": "dormitory_name",
	"category":       "category",
	"expense_date":   "expense_date",
	"amount":         "amount",
	"description":    "description",
}

const DefaultSortKey = "expense_date"

type CreateInput struct {
	DormitoryID uuid.UUID
	Category    expensedomain.ExpenseCategory
	ExpenseDate time.Time
	Amount      float64
	Description string
	CreatedBy   *uuid.UUID
}

type UpdateInput struct {
	Category    *expensedomain.ExpenseCategory
	ExpenseDate *time.Time
	Amount      *float64
	Description *string
	UpdatedBy   *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]expensedomain.Expense, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (expensedomain.Expense, error)
	Create(ctx context.Context, input CreateInput) (expensedomain.Expense, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (expensedomain.Expense, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

// ActivityLogger records who created, changed or deleted an expense for the
// audit trail. Failures to record are logged but never block the expense flow.
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
// never fail the expense flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, expense expensedomain.Expense, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	var dormitoryRef *uuid.UUID
	if expense.DormitoryID != uuid.Nil {
		dormitoryRef = &expense.DormitoryID
	}
	entityID := expense.ID
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "expense",
		EntityID:    &entityID,
		DormitoryID: dormitoryRef,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func expenseActivityDescription(expense expensedomain.Expense) string {
	return fmt.Sprintf("expense: %s, %.2f THB", expense.Category, expense.Amount)
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]expensedomain.Expense, int64, error) {
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

	expenses, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return expenses, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (expensedomain.Expense, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (expensedomain.Expense, error) {
	input.Description = strings.TrimSpace(input.Description)
	if input.Category == "" {
		input.Category = expensedomain.ExpenseCategoryOther
	}

	if input.DormitoryID == uuid.Nil || input.ExpenseDate.IsZero() {
		return expensedomain.Expense{}, expensedomain.ErrRequiredExpenseData
	}
	if input.Amount <= 0 {
		return expensedomain.Expense{}, expensedomain.ErrInvalidExpenseAmount
	}
	if !input.Category.Valid() {
		return expensedomain.Expense{}, expensedomain.ErrInvalidCategory
	}

	expense, err := s.repo.Create(ctx, input)
	if err != nil {
		return expensedomain.Expense{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", expense, "Created "+expenseActivityDescription(expense), ipAddress)
	return expense, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (expensedomain.Expense, error) {
	if input.Category != nil && !input.Category.Valid() {
		return expensedomain.Expense{}, expensedomain.ErrInvalidCategory
	}
	if input.Amount != nil && *input.Amount <= 0 {
		return expensedomain.Expense{}, expensedomain.ErrInvalidExpenseAmount
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		input.Description = &description
	}

	// Read before the update so the log can show what the amount changed from.
	// It also checks the requester's access, so a missing or out-of-scope
	// expense fails here as not found, same as the update itself would.
	before, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return expensedomain.Expense{}, err
	}

	expense, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return expensedomain.Expense{}, err
	}

	description := fmt.Sprintf("Updated expense: %s, %.2f -> %.2f THB", expense.Category, before.Amount, expense.Amount)
	s.recordActivity(ctx, &requesterID, "UPDATE", expense, description, ipAddress)
	return expense, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	// Read first: once deleted there is no category or amount left to describe.
	expense, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", expense, "Deleted "+expenseActivityDescription(expense), ipAddress)
	return nil
}
