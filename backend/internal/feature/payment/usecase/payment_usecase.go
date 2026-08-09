package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/delivery/http/apierror"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/payment/domain"

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
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return p, nil
}

func (uc *PaymentUseCase) Create(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.PaymentDetail, error) {
	var total float64
	for _, s := range req.Splits {
		total += s.Amount
	}
	method := req.Splits[0].Method
	if len(req.Splits) > 1 {
		method = domain.PaymentMethodMixed
	}

	p := &domain.Payment{
		ID:          uuid.New().String(),
		BillID:      req.BillID,
		Amount:      total,
		Method:      method,
		PaymentDate: req.PaymentDate,
		Note:        req.Note,
	}

	splits := make([]domain.PaymentSplit, len(req.Splits))
	for i, s := range req.Splits {
		splits[i] = domain.PaymentSplit{
			ID:        uuid.New().String(),
			PaymentID: p.ID,
			Method:    s.Method,
			Amount:    s.Amount,
		}
	}

	if err := uc.repo.Create(ctx, p, splits); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindDetailByID(ctx, p.ID)
}

func (uc *PaymentUseCase) Update(ctx context.Context, id string, req *domain.UpdatePaymentRequest) (*domain.PaymentDetail, error) {
	p, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	var total float64
	for _, s := range req.Splits {
		total += s.Amount
	}
	method := req.Splits[0].Method
	if len(req.Splits) > 1 {
		method = domain.PaymentMethodMixed
	}

	p.Amount = total
	p.Method = method
	p.PaymentDate = req.PaymentDate
	p.Note = req.Note

	splits := make([]domain.PaymentSplit, len(req.Splits))
	for i, s := range req.Splits {
		splits[i] = domain.PaymentSplit{
			ID:        uuid.New().String(),
			PaymentID: id,
			Method:    s.Method,
			Amount:    s.Amount,
		}
	}

	if err := uc.repo.Update(ctx, p, splits); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindDetailByID(ctx, id)
}

func (uc *PaymentUseCase) Delete(ctx context.Context, id string) error {
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
