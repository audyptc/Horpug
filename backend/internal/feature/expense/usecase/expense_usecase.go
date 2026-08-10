package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/expense/domain"

	"github.com/google/uuid"
)

type ExpenseUseCase struct {
	repo domain.ExpenseRepository
}

func NewExpenseUseCase(repo domain.ExpenseRepository) *ExpenseUseCase {
	return &ExpenseUseCase{repo: repo}
}

func (uc *ExpenseUseCase) List(ctx context.Context, limit, offset int) ([]*domain.Expense, int, error) {
	total, err := uc.repo.Count(ctx)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	list, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return list, total, nil
}

func (uc *ExpenseUseCase) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	e, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return e, nil
}

func (uc *ExpenseUseCase) Create(ctx context.Context, req *domain.CreateExpenseRequest) (*domain.Expense, error) {
	e := &domain.Expense{
		ID:          uuid.New().String(),
		ExpenseDate: req.ExpenseDate,
		Category:    req.Category,
		Description: req.Description,
		Amount:      req.Amount,
		Note:        req.Note,
	}
	if err := uc.repo.Create(ctx, e); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, e.ID)
}

func (uc *ExpenseUseCase) Update(ctx context.Context, id string, req *domain.UpdateExpenseRequest) (*domain.Expense, error) {
	e, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	e.ExpenseDate = req.ExpenseDate
	e.Category = req.Category
	e.Description = req.Description
	e.Amount = req.Amount
	e.Note = req.Note

	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindByID(ctx, id)
}

func (uc *ExpenseUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
