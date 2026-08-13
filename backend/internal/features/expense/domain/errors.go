package domain

import "errors"

var (
	ErrExpenseNotFound      = errors.New("expense not found")
	ErrRequiredExpenseData  = errors.New("dormitory_id and expense_date are required")
	ErrInvalidExpenseAmount = errors.New("amount must be greater than zero")
	ErrInvalidCategory      = errors.New("invalid expense category")
	ErrDormitoryNotFound    = errors.New("dormitory not found")
)
