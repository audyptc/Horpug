package usecase

import (
	"context"

	dashboarddomain "apihorpug/internal/features/dashboard/domain"

	"github.com/google/uuid"
)

type Repository interface {
	GetSummary(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID) (dashboarddomain.Summary, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetSummary counts over the dormitories the requester may access, or over one
// of them when dormitoryID is set.
func (s *Service) GetSummary(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID) (dashboarddomain.Summary, error) {
	return s.repo.GetSummary(ctx, requesterID, dormitoryID)
}
