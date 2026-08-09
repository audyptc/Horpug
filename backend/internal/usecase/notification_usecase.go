package usecase

import (
	"context"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"
)

type NotificationUseCase struct {
	notifRepo domain.NotificationRepository
}

func NewNotificationUseCase(notifRepo domain.NotificationRepository) *NotificationUseCase {
	return &NotificationUseCase{notifRepo: notifRepo}
}

func (uc *NotificationUseCase) List(ctx context.Context) ([]*domain.NotificationItem, error) {
	overdue, err := uc.notifRepo.ListOverdueBills(ctx, 20)
	if err != nil {
		return nil, apierror.Internal(err)
	}

	expiring, err := uc.notifRepo.ListExpiringContracts(ctx, 30, 20)
	if err != nil {
		return nil, apierror.Internal(err)
	}

	maintenance, err := uc.notifRepo.ListOpenMaintenance(ctx, 20)
	if err != nil {
		return nil, apierror.Internal(err)
	}

	result := make([]*domain.NotificationItem, 0, len(overdue)+len(expiring)+len(maintenance))
	result = append(result, overdue...)
	result = append(result, expiring...)
	result = append(result, maintenance...)
	return result, nil
}
