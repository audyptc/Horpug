package delivery

import (
	"apigofiberhorpug/internal/features/expense/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
)

func validateCreateExpenseRequest(req *domain.CreateExpenseRequest) error {
	validCategories := map[domain.ExpenseCategory]bool{
		domain.ExpenseCategoryRepair:    true,
		domain.ExpenseCategoryUtilities: true,
		domain.ExpenseCategorySupplies:  true,
		domain.ExpenseCategorySalary:    true,
		domain.ExpenseCategoryOther:     true,
	}
	if req.Description == "" {
		return apierror.BadRequest("description is required")
	}
	if req.ExpenseDate.IsZero() {
		return apierror.BadRequest("expense_date is required")
	}
	if !validCategories[req.Category] {
		return apierror.BadRequest("category must be one of: repair, utilities, supplies, salary, other")
	}
	if req.Amount <= 0 {
		return apierror.BadRequest("amount must be greater than 0")
	}
	return nil
}

func validateUpdateExpenseRequest(req *domain.UpdateExpenseRequest) error {
	validCategories := map[domain.ExpenseCategory]bool{
		domain.ExpenseCategoryRepair:    true,
		domain.ExpenseCategoryUtilities: true,
		domain.ExpenseCategorySupplies:  true,
		domain.ExpenseCategorySalary:    true,
		domain.ExpenseCategoryOther:     true,
	}
	if req.Description == "" {
		return apierror.BadRequest("description is required")
	}
	if req.ExpenseDate.IsZero() {
		return apierror.BadRequest("expense_date is required")
	}
	if !validCategories[req.Category] {
		return apierror.BadRequest("category must be one of: repair, utilities, supplies, salary, other")
	}
	if req.Amount <= 0 {
		return apierror.BadRequest("amount must be greater than 0")
	}
	return nil
}
