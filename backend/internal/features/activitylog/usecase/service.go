package usecase

import (
	"context"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"

	"github.com/google/uuid"
)

type CreateInput struct {
	UserID      *uuid.UUID
	Action      string
	EntityType  string
	EntityID    *uuid.UUID
	Description string
	IPAddress   string
}

type ListFilter struct {
	UserID     *uuid.UUID
	EntityType string
	EntityID   *uuid.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
	Limit      int
	Offset     int
}

type Repository interface {
	Count(ctx context.Context, filter ListFilter) (int64, error)
	List(ctx context.Context, filter ListFilter) ([]activitylogdomain.ActivityLog, error)
	GetByID(ctx context.Context, id uuid.UUID) (activitylogdomain.ActivityLog, error)
	Create(ctx context.Context, input CreateInput) (activitylogdomain.ActivityLog, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]activitylogdomain.ActivityLog, int64, error) {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	logs, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (activitylogdomain.ActivityLog, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (activitylogdomain.ActivityLog, error) {
	input.Action = strings.TrimSpace(input.Action)
	input.EntityType = strings.TrimSpace(input.EntityType)
	input.Description = strings.TrimSpace(input.Description)
	return s.repo.Create(ctx, input)
}
