package domain

import (
	"context"
	"time"
)

type ExpenseCategory string

const (
	ExpenseCategoryRepair    ExpenseCategory = "repair"
	ExpenseCategoryUtilities ExpenseCategory = "utilities"
	ExpenseCategorySupplies  ExpenseCategory = "supplies"
	ExpenseCategorySalary    ExpenseCategory = "salary"
	ExpenseCategoryOther     ExpenseCategory = "other"
)

type Expense struct {
	ID          string          `json:"id"`
	ExpenseDate time.Time       `json:"expense_date"`
	Category    ExpenseCategory `json:"category"`
	Description string          `json:"description"`
	Amount      float64         `json:"amount"`
	Note        string          `json:"note"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type CreateExpenseRequest struct {
	ExpenseDate time.Time       `json:"expense_date"`
	Category    ExpenseCategory `json:"category"`
	Description string          `json:"description"`
	Amount      float64         `json:"amount"`
	Note        string          `json:"note"`
}

type UpdateExpenseRequest struct {
	ExpenseDate time.Time       `json:"expense_date"`
	Category    ExpenseCategory `json:"category"`
	Description string          `json:"description"`
	Amount      float64         `json:"amount"`
	Note        string          `json:"note"`
}

type ExpenseRepository interface {
	FindByID(ctx context.Context, id string) (*Expense, error)
	List(ctx context.Context, limit, offset int) ([]*Expense, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, e *Expense) error
	Update(ctx context.Context, e *Expense) error
	Delete(ctx context.Context, id string) error
}
