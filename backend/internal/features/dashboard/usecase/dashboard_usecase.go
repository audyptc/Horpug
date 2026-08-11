package usecase

import (
	"context"

	"apigofiberhorpug/internal/features/dashboard/domain"
)

type DashboardUseCase struct {
	repo domain.DashboardRepository
}

func NewDashboardUseCase(repo domain.DashboardRepository) *DashboardUseCase {
	return &DashboardUseCase{repo: repo}
}

func (uc *DashboardUseCase) Summary(ctx context.Context, dormitoryID string) (*domain.DashboardSummary, error) {
	return uc.repo.Summary(ctx, dormitoryID)
}
