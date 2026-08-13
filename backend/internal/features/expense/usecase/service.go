package usecase

import (
	"context"
	"strings"
	"time"

	expensedomain "apihorpug/internal/features/expense/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	DormitoryID *uuid.UUID
	Category    *expensedomain.ExpenseCategory
	DateFrom    *time.Time
	DateTo      *time.Time
}

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

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]expensedomain.Expense, int64, error) {
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

func (s *Service) Create(ctx context.Context, input CreateInput) (expensedomain.Expense, error) {
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

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (expensedomain.Expense, error) {
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

	return s.repo.Update(ctx, id, requesterID, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
