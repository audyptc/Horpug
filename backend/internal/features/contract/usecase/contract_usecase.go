package usecase

import (
	"context"
	"errors"

	"apigofiberhorpug/internal/features/contract/domain"
	roomdomain "apigofiberhorpug/internal/features/room/domain"
	tenantdomain "apigofiberhorpug/internal/features/tenant/domain"
	coredomain "apigofiberhorpug/internal/shared/domain"
	"apigofiberhorpug/internal/shared/http/apierror"

	"github.com/google/uuid"
)

type ContractUseCase struct {
	contractRepo domain.ContractRepository
	roomRepo     roomdomain.RoomRepository
	tenantRepo   tenantdomain.TenantRepository
}

func NewContractUseCase(
	contractRepo domain.ContractRepository,
	roomRepo roomdomain.RoomRepository,
	tenantRepo tenantdomain.TenantRepository,
) *ContractUseCase {
	return &ContractUseCase{contractRepo: contractRepo, roomRepo: roomRepo, tenantRepo: tenantRepo}
}

func (uc *ContractUseCase) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.ContractDetail, int, error) {
	total, err := uc.contractRepo.Count(ctx, dormitoryID)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	list, err := uc.contractRepo.List(ctx, dormitoryID, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return list, total, nil
}

func (uc *ContractUseCase) GetByID(ctx context.Context, dormitoryID, id string) (*domain.ContractDetail, error) {
	d, err := uc.contractRepo.FindDetailByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}
	return d, nil
}

func (uc *ContractUseCase) Create(ctx context.Context, dormitoryID string, req *domain.CreateContractRequest, actorID string) (*domain.ContractDetail, error) {
	if _, err := uc.tenantRepo.FindByID(ctx, dormitoryID, req.TenantID); err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound("tenant not found")
		}
		return nil, apierror.Internal(err)
	}

	room, err := uc.roomRepo.FindByID(ctx, dormitoryID, req.RoomID)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound("room not found")
		}
		return nil, apierror.Internal(err)
	}
	if room.Status != "available" {
		return nil, apierror.BadRequest("room is not available")
	}

	numOccupants := req.NumOccupants
	if numOccupants < 1 {
		numOccupants = 1
	}
	c := &domain.Contract{
		ID:           uuid.New().String(),
		TenantID:     req.TenantID,
		RoomID:       req.RoomID,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		RentPrice:    req.RentPrice,
		Deposit:      req.Deposit,
		NumOccupants: numOccupants,
		Status:       domain.ContractStatusActive,
		Note:         req.Note,
		CreatedBy:    actorID,
		UpdatedBy:    actorID,
	}
	if err := uc.contractRepo.Create(ctx, c); err != nil {
		return nil, apierror.Internal(err)
	}

	room.Status = "occupied"
	if err := uc.roomRepo.Update(ctx, room); err != nil {
		return nil, apierror.Internal(err)
	}

	return uc.contractRepo.FindDetailByID(ctx, dormitoryID, c.ID)
}

func (uc *ContractUseCase) Update(ctx context.Context, dormitoryID, id string, req *domain.UpdateContractRequest, actorID string) (*domain.ContractDetail, error) {
	c, err := uc.contractRepo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return nil, apierror.NotFound(err.Error())
		}
		return nil, apierror.Internal(err)
	}

	prevStatus := c.Status

	c.EndDate = req.EndDate
	if req.RentPrice > 0 {
		c.RentPrice = req.RentPrice
	}
	if req.Deposit >= 0 {
		c.Deposit = req.Deposit
	}
	if req.Status != "" {
		c.Status = req.Status
	}
	if req.NumOccupants != nil && *req.NumOccupants >= 1 {
		c.NumOccupants = *req.NumOccupants
	}
	c.Note = req.Note
	c.UpdatedBy = actorID

	if err := uc.contractRepo.Update(ctx, dormitoryID, c); err != nil {
		return nil, apierror.Internal(err)
	}

	// When contract moves from active -> terminated/expired, free the room
	if prevStatus == domain.ContractStatusActive &&
		(c.Status == domain.ContractStatusTerminated || c.Status == domain.ContractStatusExpired) {
		room, err := uc.roomRepo.FindByID(ctx, dormitoryID, c.RoomID)
		if err == nil {
			room.Status = "available"
			_ = uc.roomRepo.Update(ctx, room)
		}
	}

	return uc.contractRepo.FindDetailByID(ctx, dormitoryID, id)
}

func (uc *ContractUseCase) Delete(ctx context.Context, dormitoryID, id string) error {
	c, err := uc.contractRepo.FindByID(ctx, dormitoryID, id)
	if err != nil {
		if errors.Is(err, coredomain.ErrNotFound) {
			return apierror.NotFound(err.Error())
		}
		return apierror.Internal(err)
	}

	if err := uc.contractRepo.Delete(ctx, dormitoryID, id); err != nil {
		return apierror.Internal(err)
	}

	// Free the room if contract was active
	if c.Status == domain.ContractStatusActive {
		room, err := uc.roomRepo.FindByID(ctx, dormitoryID, c.RoomID)
		if err == nil {
			room.Status = "available"
			_ = uc.roomRepo.Update(ctx, room)
		}
	}
	return nil
}
