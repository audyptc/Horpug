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
	DormitoryID string          `json:"dormitory_id"`
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
	FindByID(ctx context.Context, dormitoryID, id string) (*Expense, error)
	List(ctx context.Context, dormitoryID string, limit, offset int) ([]*Expense, error)
	Count(ctx context.Context, dormitoryID string) (int, error)
	Create(ctx context.Context, e *Expense) error
	Update(ctx context.Context, e *Expense) error
	Delete(ctx context.Context, dormitoryID, id string) error
}
