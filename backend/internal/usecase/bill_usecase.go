package usecase

import (
	"context"
	"errors"
	"time"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"
	contractdomain "apigofiberhorpug/internal/feature/contract/domain"

	"github.com/google/uuid"
)

type BillUseCase struct {
	billRepo     domain.BillRepository
	contractRepo contractdomain.ContractRepository
}

func NewBillUseCase(billRepo domain.BillRepository, contractRepo contractdomain.ContractRepository) *BillUseCase {
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

func (uc *BillUseCase) Create(ctx context.Context, req *domain.CreateBillRequest, actorID string) (*domain.BillDetail, error) {
	if _, err := uc.contractRepo.FindByID(ctx, req.ContractID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound("contract not found")
		}
		return nil, apierror.Internal(err)
	}

	var otherAmount float64
	for _, item := range req.OtherItems {
		otherAmount += item.Amount
	}

	b := &domain.Bill{
		ID:             uuid.New().String(),
		ContractID:     req.ContractID,
		BillingMonth:   req.BillingMonth,
		RentAmount:     req.RentAmount,
		ElectricAmount: req.ElectricAmount,
		WaterAmount:    req.WaterAmount,
		ParkingAmount:  req.ParkingAmount,
		OtherAmount:    otherAmount,
		TotalAmount:    req.RentAmount + req.ElectricAmount + req.WaterAmount + req.ParkingAmount + otherAmount,
		Status:         domain.BillStatusUnpaid,
		DueDate:        req.DueDate,
		Note:           req.Note,
		CreatedBy:      actorID,
		UpdatedBy:      actorID,
	}
	if err := uc.billRepo.Create(ctx, b); err != nil {
		return nil, apierror.Internal(err)
	}

	items := make([]domain.BillOtherItem, len(req.OtherItems))
	for i, inp := range req.OtherItems {
		items[i] = domain.BillOtherItem{
			ID:        uuid.New().String(),
			BillID:    b.ID,
			Label:     inp.Label,
			Amount:    inp.Amount,
			SortOrder: i,
		}
	}
	if err := uc.billRepo.ReplaceOtherItems(ctx, b.ID, items); err != nil {
		return nil, apierror.Internal(err)
	}

	return uc.billRepo.FindDetailByID(ctx, b.ID)
}

func (uc *BillUseCase) Update(ctx context.Context, id string, req *domain.UpdateBillRequest, actorID string) (*domain.BillDetail, error) {
	b, err := uc.billRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	var otherAmount float64
	if req.OtherItems != nil {
		for _, item := range *req.OtherItems {
			otherAmount += item.Amount
		}
	} else {
		otherAmount = b.OtherAmount
	}

	b.RentAmount = req.RentAmount
	b.ElectricAmount = req.ElectricAmount
	b.WaterAmount = req.WaterAmount
	b.ParkingAmount = req.ParkingAmount
	b.OtherAmount = otherAmount
	b.TotalAmount = req.RentAmount + req.ElectricAmount + req.WaterAmount + req.ParkingAmount + otherAmount
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

	b.UpdatedBy = actorID

	if err := uc.billRepo.Update(ctx, b); err != nil {
		return nil, apierror.Internal(err)
	}

	if req.OtherItems != nil {
		items := make([]domain.BillOtherItem, len(*req.OtherItems))
		for i, inp := range *req.OtherItems {
			items[i] = domain.BillOtherItem{
				ID:        uuid.New().String(),
				BillID:    id,
				Label:     inp.Label,
				Amount:    inp.Amount,
				SortOrder: i,
			}
		}
		if err := uc.billRepo.ReplaceOtherItems(ctx, id, items); err != nil {
			return nil, apierror.Internal(err)
		}
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
