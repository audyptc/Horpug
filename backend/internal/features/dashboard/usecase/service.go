package usecase

import (
	"context"

	dashboarddomain "apihorpug/internal/features/dashboard/domain"

	"github.com/google/uuid"
)

type Repository interface {
	GetSummary(ctx context.Context, requesterID uuid.UUID) (dashboarddomain.Summary, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetSummary(ctx context.Context, requesterID uuid.UUID) (dashboarddomain.Summary, error) {
	return s.repo.GetSummary(ctx, requesterID)
}
