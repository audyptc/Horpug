package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo domain.PaymentRepository
}

func NewPaymentUseCase(repo domain.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

func (uc *PaymentUseCase) List(ctx context.Context, limit, offset int) ([]*domain.PaymentDetail, int, error) {
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

func (uc *PaymentUseCase) GetByID(ctx context.Context, id string) (*domain.PaymentDetail, error) {
	p, err := uc.repo.FindDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return p, nil
}

func (uc *PaymentUseCase) Create(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.PaymentDetail, error) {
	p := &domain.Payment{
		ID:          uuid.New().String(),
		BillID:      req.BillID,
		Amount:      req.Amount,
		Method:      req.Method,
		PaymentDate: req.PaymentDate,
		Note:        req.Note,
	}
	if err := uc.repo.Create(ctx, p); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindDetailByID(ctx, p.ID)
}

func (uc *PaymentUseCase) Update(ctx context.Context, id string, req *domain.UpdatePaymentRequest) (*domain.PaymentDetail, error) {
	p, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	p.Amount = req.Amount
	p.Method = req.Method
	p.PaymentDate = req.PaymentDate
	p.Note = req.Note

	if err := uc.repo.Update(ctx, p); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindDetailByID(ctx, id)
}

func (uc *PaymentUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
