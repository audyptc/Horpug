package domain

import (
	"time"

	"github.com/google/uuid"
)

type ExpenseCategory string

const (
	ExpenseCategoryMaintenance ExpenseCategory = "maintenance"
	ExpenseCategoryUtility     ExpenseCategory = "utility"
	ExpenseCategorySalary      ExpenseCategory = "salary"
	ExpenseCategorySupplies    ExpenseCategory = "supplies"
	ExpenseCategoryOther       ExpenseCategory = "other"
)

func (c ExpenseCategory) Valid() bool {
	switch c {
	case ExpenseCategoryMaintenance, ExpenseCategoryUtility, ExpenseCategorySalary, ExpenseCategorySupplies, ExpenseCategoryOther:
		return true
	}
	return false
}

// Expense is an operating cost the dormitory itself incurs (e.g. maintenance,
// staff salary, supplies), as opposed to a charge billed to a tenant.
type Expense struct {
	ID            uuid.UUID       `json:"id"`
	DormitoryID   uuid.UUID       `json:"dormitory_id"`
	DormitoryName string          `json:"dormitory_name,omitempty"`
	Category      ExpenseCategory `json:"category"`
	ExpenseDate   time.Time       `json:"expense_date"`
	Amount        float64         `json:"amount"`
	Description   string          `json:"description"`
	CreatedBy     *uuid.UUID      `json:"created_by,omitempty"`
	UpdatedBy     *uuid.UUID      `json:"updated_by,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}
