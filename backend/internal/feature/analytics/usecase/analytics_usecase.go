package usecase

import (
	"context"

	"apigofiberhorpug/internal/feature/analytics/domain"
)

type AnalyticsUseCase struct {
	repo domain.AnalyticsRepository
}

func NewAnalyticsUseCase(repo domain.AnalyticsRepository) *AnalyticsUseCase {
	return &AnalyticsUseCase{repo: repo}
}

func (uc *AnalyticsUseCase) Summary(ctx context.Context, months int) (*domain.AnalyticsSummary, error) {
	if months <= 0 || months > 24 {
		months = 12
	}
	return uc.repo.Summary(ctx, months)
}
