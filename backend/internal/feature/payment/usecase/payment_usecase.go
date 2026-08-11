package usecase

import (
	"context"
	"errors"

	billdomain "apigofiberhorpug/internal/feature/bill/domain"
	"apigofiberhorpug/internal/feature/payment/domain"
	coredomain "apigofiberhorpug/internal/shared/domain"
	"apigofiberhorpug/internal/shared/http/apierror"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo     domain.PaymentRepository
	billRepo billdomain.BillRepository
}

func NewPaymentUseCase(repo domain.PaymentRepository, billRepo billdomain.BillRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo, billRepo: billRepo}
}

func (uc *PaymentUseCase) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.PaymentDetail, int, error) {
	total, err := uc.repo.Count(ctx, dormitoryID)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	list, err := uc.repo.List(ctx, dormitoryID, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return list, total, nil
}

func (uc *PaymentUseCase) GetByID(ctx context.Context, dormitoryID, id string) (*domain.PaymentDetail, error) {
	p, err := uc.repo.FindDetailByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return p, nil
}

func (uc *PaymentUseCase) Create(ctx context.Context, dormitoryID string, req *domain.CreatePaymentRequest) (*domain.PaymentDetail, error) {
	if _, err := uc.billRepo.FindByID(ctx, dormitoryID, req.BillID); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound("bill not found")
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
	return uc.repo.FindDetailByID(ctx, dormitoryID, p.ID)
}

func (uc *PaymentUseCase) Update(ctx context.Context, dormitoryID, id string, req *domain.UpdatePaymentRequest) (*domain.PaymentDetail, error) {
	p, err := uc.repo.FindByID(ctx, dormitoryID, id)
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

	if err := uc.repo.Update(ctx, dormitoryID, p, splits); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.repo.FindDetailByID(ctx, dormitoryID, id)
}

func (uc *PaymentUseCase) Delete(ctx context.Context, dormitoryID, id string) error {
	if _, err := uc.repo.FindByID(ctx, dormitoryID, id); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.repo.Delete(ctx, dormitoryID, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
