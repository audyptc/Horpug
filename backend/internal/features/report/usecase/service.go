package usecase

import (
	"context"
	"errors"
	"math"

	reportdomain "apihorpug/internal/features/report/domain"

	"github.com/google/uuid"
)

var ErrInvalidPeriod = errors.New("year and month (1-12) are required")

type Repository interface {
	Monthly(ctx context.Context, f reportdomain.Filter) (reportdomain.MonthlyReport, error)
	Details(ctx context.Context, f reportdomain.Filter) (reportdomain.Details, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func filter(requesterID uuid.UUID, dormitoryID *uuid.UUID, year, month int) (reportdomain.Filter, error) {
	if year < 2000 || year > 3000 || month < 1 || month > 12 {
		return reportdomain.Filter{}, ErrInvalidPeriod
	}
	return reportdomain.Filter{RequesterID: requesterID, DormitoryID: dormitoryID, Year: year, Month: month}, nil
}

// Monthly builds the month's report over the dormitories the requester
// manages, or just dormitoryID when given.
func (s *Service) Monthly(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, year, month int) (reportdomain.MonthlyReport, error) {
	f, err := filter(requesterID, dormitoryID, year, month)
	if err != nil {
		return reportdomain.MonthlyReport{}, err
	}
	rep, err := s.repo.Monthly(ctx, f)
	if err != nil {
		return reportdomain.MonthlyReport{}, err
	}

	rep.Net = round2(rep.Income.Total - rep.Expenses.Total)
	for i := range rep.Dormitories {
		rep.Dormitories[i].Net = round2(rep.Dormitories[i].Income - rep.Dormitories[i].Expenses)
	}
	return rep, nil
}

// Export builds the month's report as an Excel workbook.
func (s *Service) Export(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, year, month int) ([]byte, error) {
	rep, err := s.Monthly(ctx, requesterID, dormitoryID, year, month)
	if err != nil {
		return nil, err
	}
	f, err := filter(requesterID, dormitoryID, year, month)
	if err != nil {
		return nil, err
	}
	details, err := s.repo.Details(ctx, f)
	if err != nil {
		return nil, err
	}
	return buildWorkbook(rep, details)
}
