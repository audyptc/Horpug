package usecase

import (
	"context"
	"errors"
	"time"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"

	"github.com/google/uuid"
)

type BillUseCase struct {
	billRepo     domain.BillRepository
	contractRepo domain.ContractRepository
}

func NewBillUseCase(billRepo domain.BillRepository, contractRepo domain.ContractRepository) *BillUseCase {
	return &BillUseCase{billRepo: billRepo, contractRepo: contractRepo}
}

func (uc *BillUseCase) List(ctx context.Context, limit, offset int) ([]*domain.BillDetail, int, error) {
	total, err := uc.billRepo.Count(ctx)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	list, err := uc.billRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return list, total, nil
}

func (uc *BillUseCase) GetByID(ctx context.Context, id string) (*domain.BillDetail, error) {
	d, err := uc.billRepo.FindDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return d, nil
}

func (uc *BillUseCase) Create(ctx context.Context, req *domain.CreateBillRequest) (*domain.BillDetail, error) {
	if _, err := uc.contractRepo.FindByID(ctx, req.ContractID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound("contract not found")
		}
		return nil, apierror.Internal(err)
	}

	total := req.RentAmount + req.ElectricAmount + req.WaterAmount + req.OtherAmount

	b := &domain.Bill{
		ID:             uuid.New().String(),
		ContractID:     req.ContractID,
		BillingMonth:   req.BillingMonth,
		RentAmount:     req.RentAmount,
		ElectricAmount: req.ElectricAmount,
		WaterAmount:    req.WaterAmount,
		OtherAmount:    req.OtherAmount,
		OtherNote:      req.OtherNote,
		TotalAmount:    total,
		Status:         domain.BillStatusUnpaid,
		DueDate:        req.DueDate,
		Note:           req.Note,
	}
	if err := uc.billRepo.Create(ctx, b); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.billRepo.FindDetailByID(ctx, b.ID)
}

func (uc *BillUseCase) Update(ctx context.Context, id string, req *domain.UpdateBillRequest) (*domain.BillDetail, error) {
	b, err := uc.billRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	b.RentAmount = req.RentAmount
	b.ElectricAmount = req.ElectricAmount
	b.WaterAmount = req.WaterAmount
	b.OtherAmount = req.OtherAmount
	b.OtherNote = req.OtherNote
	b.TotalAmount = req.RentAmount + req.ElectricAmount + req.WaterAmount + req.OtherAmount
	b.DueDate = req.DueDate
	b.Note = req.Note

	if req.Status != "" {
		prevStatus := b.Status
		b.Status = req.Status
		if prevStatus != domain.BillStatusPaid && req.Status == domain.BillStatusPaid {
			now := time.Now()
			b.PaidAt = &now
		} else if req.Status != domain.BillStatusPaid {
			b.PaidAt = nil
		}
	}

	if err := uc.billRepo.Update(ctx, b); err != nil {
		return nil, apierror.Internal(err)
	}
	return uc.billRepo.FindDetailByID(ctx, id)
}

func (uc *BillUseCase) Delete(ctx context.Context, id string) error {
	if _, err := uc.billRepo.FindByID(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}
	if err := uc.billRepo.Delete(ctx, id); err != nil {
		return apierror.Internal(err)
	}
	return nil
}
