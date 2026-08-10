package usecase

import (
	"context"
	"time"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/report/domain"
)

type ReportUseCase struct {
	repo domain.ReportRepository
}

func NewReportUseCase(repo domain.ReportRepository) *ReportUseCase {
	return &ReportUseCase{repo: repo}
}

func (uc *ReportUseCase) IncomeReport(ctx context.Context, dormitoryID, from, to string) (*domain.IncomeReport, error) {
	from, to, err := normalizeDateRange(from, to)
	if err != nil {
		return nil, err
	}
	return uc.repo.IncomeReport(ctx, dormitoryID, from, to)
}

func (uc *ReportUseCase) ExpenseReport(ctx context.Context, dormitoryID, from, to string) (*domain.ExpenseReport, error) {
	from, to, err := normalizeDateRange(from, to)
	if err != nil {
		return nil, err
	}
	return uc.repo.ExpenseReport(ctx, dormitoryID, from, to)
}

func (uc *ReportUseCase) OccupancyReport(ctx context.Context, dormitoryID string) (*domain.OccupancyReport, error) {
	return uc.repo.OccupancyReport(ctx, dormitoryID)
}

func normalizeDateRange(from, to string) (string, string, error) {
	now := time.Now()
	if from == "" {
		from = now.AddDate(0, -5, 0).Format("2006-01-02")
	}
	if to == "" {
		to = now.Format("2006-01-02")
	}
	fromT, err := time.Parse("2006-01", from)
	if err != nil {
		fromT, err = time.Parse("2006-01-02", from)
		if err != nil {
			return "", "", apierror.BadRequest("invalid from date, expected YYYY-MM or YYYY-MM-DD")
		}
	}
	toT, err := time.Parse("2006-01", to)
	if err != nil {
		toT, err = time.Parse("2006-01-02", to)
		if err != nil {
			return "", "", apierror.BadRequest("invalid to date, expected YYYY-MM or YYYY-MM-DD")
		}
	}
	if fromT.After(toT) {
		return "", "", apierror.BadRequest("from date must be before or equal to to date")
	}
	return fromT.Format("2006-01-02"), toT.Format("2006-01-02"), nil
}
