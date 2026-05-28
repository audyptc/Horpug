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
	items, err := uc.notifRepo.ListOverdueBills(ctx, 20)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	return items, nil
}
