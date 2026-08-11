package usecase

import (
	"context"
	"strings"

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
	Limit      int
	Offset     int
}

type Repository interface {
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

func (s *Service) List(ctx context.Context, filter ListFilter) ([]activitylogdomain.ActivityLog, error) {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.List(ctx, filter)
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
