package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	documentdomain "apihorpug/internal/features/document/domain"
	"apihorpug/internal/platform/filestore"

	"github.com/google/uuid"
)

type ListFilters struct {
	DormitoryID *uuid.UUID
	TenantID    *uuid.UUID
	RoomID      *uuid.UUID
	Category    *documentdomain.DocumentCategory

	Search string
	// Columns narrows individual columns by substring, keyed by FilterColumns.
	// Search casts a wide OR across name, dormitory, tenant and room; these
	// are ANDed on top, so the two answer different questions and compose.
	Columns  map[string]string
	SortKey  string
	SortDesc bool
}

// FilterColumns whitelists the columns a caller may match a substring against,
// passed as f[<column>]=value. The repository resolves each key to the actual
// SQL it needs — this map exists purely so an unrecognised column is rejected
// rather than silently ignored.
var FilterColumns = map[string]string{
	"name":           "name",
	"dormitory_name": "dormitory_name",
	"tenant_name":    "tenant_name",
	"room_number":    "room_number",
}

// SortColumns whitelists the sort keys the API accepts. As with FilterColumns,
// the repository resolves each key to its actual ORDER BY expression.
var SortColumns = map[string]string{
	"name":           "name",
	"category":       "category",
	"dormitory_name": "dormitory_name",
	"tenant_name":    "tenant_name",
	"room_number":    "room_number",
	"uploaded_date":  "uploaded_date",
}

const DefaultSortKey = "uploaded_date"

type CreateInput struct {
	DormitoryID uuid.UUID
	TenantID    *uuid.UUID
	RoomID      *uuid.UUID
	Name        string
	Category    documentdomain.DocumentCategory
	FileURL     string
	// Upload, when set, supplies the document's file and takes precedence over
	// FileURL. File is filled in by the service once the upload is stored.
	Upload       *FileUpload
	File         *StoredFile
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
	Upload       *FileUpload
	File         *StoredFile
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

// ActivityLogger records who uploaded, changed or deleted a document for the
// audit trail. Failures to record are logged but never block the flow.
type ActivityLogger interface {
	Create(ctx context.Context, input activitylogusecase.CreateInput) (activitylogdomain.ActivityLog, error)
}

// FileStorage holds the bytes of uploaded documents. Keys are opaque and
// returned by Save.
type FileStorage interface {
	Save(folder, ext string, r io.Reader) (string, error)
	Read(key string) ([]byte, error)
	Delete(key string) error
}

type Service struct {
	repo        Repository
	files       FileStorage
	activityLog ActivityLogger
}

func New(repo Repository, files FileStorage, activityLog ActivityLogger) *Service {
	return &Service{repo: repo, files: files, activityLog: activityLog}
}

// discardFile removes a stored file that is no longer referenced. Best-effort:
// a leftover file is wasted space, not a reason to fail the request.
func (s *Service) discardFile(key string) {
	if key == "" {
		return
	}
	if err := s.files.Delete(key); err != nil {
		log.Printf("failed to delete stored file %q: %v", key, err)
	}
}

// recordActivity is best-effort: a failure to write the audit trail must
// never fail the document flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, document documentdomain.Document, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	entityID := document.ID
	dormitoryID := document.DormitoryID
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "document",
		EntityID:    &entityID,
		DormitoryID: &dormitoryID,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func documentActivityDescription(document documentdomain.Document) string {
	return fmt.Sprintf("document: %s, %s", document.Name, document.Category)
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]documentdomain.Document, int64, error) {
	filters.Search = strings.TrimSpace(filters.Search)
	if _, ok := SortColumns[filters.SortKey]; !ok {
		filters.SortKey = DefaultSortKey
	}

	columns := make(map[string]string, len(filters.Columns))
	for key, value := range filters.Columns {
		if _, ok := FilterColumns[key]; !ok {
			continue
		}
		if value = strings.TrimSpace(value); value != "" {
			columns[key] = value
		}
	}
	filters.Columns = columns

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

// GetFile returns the document together with the bytes of its stored file.
// Access is checked by GetByID, so a document outside the requester's
// dormitories is reported as not found.
func (s *Service) GetFile(ctx context.Context, id, requesterID uuid.UUID) (documentdomain.Document, []byte, error) {
	document, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return documentdomain.Document{}, nil, err
	}
	if !document.HasFile {
		return documentdomain.Document{}, nil, documentdomain.ErrFileNotFound
	}

	data, err := s.files.Read(document.FilePath)
	if err != nil {
		if errors.Is(err, filestore.ErrNotFound) {
			return documentdomain.Document{}, nil, documentdomain.ErrFileNotFound
		}
		return documentdomain.Document{}, nil, err
	}
	return document, data, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (documentdomain.Document, error) {
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

	if input.DormitoryID == uuid.Nil || input.Name == "" || (input.FileURL == "" && input.Upload == nil) {
		return documentdomain.Document{}, documentdomain.ErrRequiredDocumentData
	}
	if !input.Category.Valid() {
		return documentdomain.Document{}, documentdomain.ErrInvalidDocumentCategory
	}

	if input.Upload != nil {
		stored, err := s.storeUpload(input.DormitoryID.String(), input.Upload)
		if err != nil {
			return documentdomain.Document{}, err
		}
		input.File = stored
		input.FileURL = ""
	}

	document, err := s.repo.Create(ctx, input)
	if err != nil {
		if input.File != nil {
			s.discardFile(input.File.Path)
		}
		return documentdomain.Document{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", document, "Created "+documentActivityDescription(document), ipAddress)
	return document, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (documentdomain.Document, error) {
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

	// Read first: it checks the requester's access before anything is written
	// to storage, and remembers the previous file so it can be cleaned up.
	previous, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return documentdomain.Document{}, err
	}

	if input.Upload != nil {
		stored, err := s.storeUpload(previous.DormitoryID.String(), input.Upload)
		if err != nil {
			return documentdomain.Document{}, err
		}
		input.File = stored
		input.FileURL = nil
	}

	document, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		if input.File != nil {
			s.discardFile(input.File.Path)
		}
		return documentdomain.Document{}, err
	}

	if previous.FilePath != "" && previous.FilePath != document.FilePath {
		s.discardFile(previous.FilePath)
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", document, "Updated "+documentActivityDescription(document), ipAddress)
	return document, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	// Read first: once deleted there is no name left to describe. It also
	// checks the requester's access, so a missing or out-of-scope document
	// fails here as not found, same as the delete itself would.
	document, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.discardFile(document.FilePath)
	s.recordActivity(ctx, &requesterID, "DELETE", document, "Deleted "+documentActivityDescription(document), ipAddress)
	return nil
}
