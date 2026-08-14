package usecase

import (
	"context"
	"strings"
	"time"

	documentdomain "apihorpug/internal/features/document/domain"

	"github.com/google/uuid"
)

type ListFilters struct {
	DormitoryID *uuid.UUID
	TenantID    *uuid.UUID
	RoomID      *uuid.UUID
	Category    *documentdomain.DocumentCategory
}

type CreateInput struct {
	DormitoryID  uuid.UUID
	TenantID     *uuid.UUID
	RoomID       *uuid.UUID
	Name         string
	Category     documentdomain.DocumentCategory
	FileURL      string
	UploadedDate time.Time
	Note         string
	CreatedBy    *uuid.UUID
}

type UpdateInput struct {
	TenantID     *uuid.UUID
	RoomID       *uuid.UUID
	Name         *string
	Category     *documentdomain.DocumentCategory
	FileURL      *string
	UploadedDate *time.Time
	Note         *string
	UpdatedBy    *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]documentdomain.Document, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (documentdomain.Document, error)
	Create(ctx context.Context, input CreateInput) (documentdomain.Document, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (documentdomain.Document, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]documentdomain.Document, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	documents, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (documentdomain.Document, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (documentdomain.Document, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.FileURL = strings.TrimSpace(input.FileURL)
	input.Note = strings.TrimSpace(input.Note)
	if input.Category == "" {
		input.Category = documentdomain.DocumentCategoryOther
	}
	if input.UploadedDate.IsZero() {
		input.UploadedDate = time.Now()
	}
	if input.TenantID != nil && *input.TenantID == uuid.Nil {
		input.TenantID = nil
	}
	if input.RoomID != nil && *input.RoomID == uuid.Nil {
		input.RoomID = nil
	}

	if input.DormitoryID == uuid.Nil || input.Name == "" || input.FileURL == "" {
		return documentdomain.Document{}, documentdomain.ErrRequiredDocumentData
	}
	if !input.Category.Valid() {
		return documentdomain.Document{}, documentdomain.ErrInvalidDocumentCategory
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (documentdomain.Document, error) {
	if input.Category != nil && !input.Category.Valid() {
		return documentdomain.Document{}, documentdomain.ErrInvalidDocumentCategory
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return documentdomain.Document{}, documentdomain.ErrRequiredDocumentData
		}
		input.Name = &name
	}
	if input.FileURL != nil {
		fileURL := strings.TrimSpace(*input.FileURL)
		if fileURL == "" {
			return documentdomain.Document{}, documentdomain.ErrRequiredDocumentData
		}
		input.FileURL = &fileURL
	}
	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		input.Note = &note
	}

	return s.repo.Update(ctx, id, requesterID, input)
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	return s.repo.Delete(ctx, id, requesterID)
}
