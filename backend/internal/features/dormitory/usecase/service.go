package usecase

import (
	"context"
	"strings"

	dormdomain "apihorpug/internal/features/dormitory/domain"

	"github.com/google/uuid"
)

type CreateInput struct {
	Name        string
	Address     string
	Phone       string
	Description string
	IsActive    bool
	ManagerIDs  []uuid.UUID
	CreatedBy   *uuid.UUID
}

type UpdateInput struct {
	Name        *string
	Address     *string
	Phone       *string
	Description *string
	IsActive    *bool
	ManagerIDs  *[]uuid.UUID
	UpdatedBy   *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, limit, offset int) ([]dormdomain.Dormitory, error)
	ListActive(ctx context.Context, requesterID uuid.UUID, search string, limit int) ([]dormdomain.Dormitory, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (dormdomain.Dormitory, error)
	Create(ctx context.Context, input CreateInput) (dormdomain.Dormitory, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (dormdomain.Dormitory, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, limit, offset int) ([]dormdomain.Dormitory, int64, error) {
	total, err := s.repo.Count(ctx, requesterID)
	if err != nil {
		return nil, 0, err
	}

	dormitories, err := s.repo.List(ctx, requesterID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return dormitories, total, nil
}

func (s *Service) ListActive(ctx context.Context, requesterID uuid.UUID, search string, limit int) ([]dormdomain.Dormitory, error) {
	return s.repo.ListActive(ctx, requesterID, strings.TrimSpace(search), limit)
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (dormdomain.Dormitory, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (dormdomain.Dormitory, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Address = strings.TrimSpace(input.Address)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Description = strings.TrimSpace(input.Description)
	if input.Name == "" {
		return dormdomain.Dormitory{}, dormdomain.ErrRequiredDormitoryData
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (dormdomain.Dormitory, error) {
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return dormdomain.Dormitory{}, dormdomain.ErrRequiredDormitoryData
		}
		input.Name = &name
	}
	if input.Address != nil {
		address := strings.TrimSpace(*input.Address)
		input.Address = &address
	}
	if input.Phone != nil {
		phone := strings.TrimSpace(*input.Phone)
		input.Phone = &phone
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		input.Description = &description
	}

	return s.repo.Update(ctx, id, requesterID, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
