package usecase

import (
	"context"
	"encoding/json"
	"time"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"

	"github.com/google/uuid"
)

type ActivityLogUseCase struct {
	repo domain.ActivityLogRepository
}

func NewActivityLogUseCase(repo domain.ActivityLogRepository) *ActivityLogUseCase {
	return &ActivityLogUseCase{repo: repo}
}

// Log บันทึก activity เงียบ ๆ (ไม่ return error เพื่อไม่ให้กระทบ main operation)
func (uc *ActivityLogUseCase) Log(ctx context.Context, actorID string, action domain.ActivityAction, entityType, entityID string, newValue interface{}) {
	var raw json.RawMessage
	if newValue != nil {
		if b, err := json.Marshal(newValue); err == nil {
			raw = b
		}
	}

	var aid *string
	if actorID != "" {
		aid = &actorID
	}

	now := time.Now()
	_ = uc.repo.Create(ctx, &domain.ActivityLog{
		ID:         uuid.New().String(),
		ActorID:    aid,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		NewValue:   raw,
		CreatedAt:  now,
	})
}

func (uc *ActivityLogUseCase) List(ctx context.Context, filter domain.ActivityLogFilter, limit, offset int) ([]*domain.ActivityLog, int, error) {
	total, err := uc.repo.Count(ctx, filter)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	list, err := uc.repo.List(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, apierror.Internal(err)
	}
	return list, total, nil
}
