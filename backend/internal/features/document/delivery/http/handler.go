package http

import (
	"context"
	"errors"
	"strings"
	"time"

	documentdomain "apihorpug/internal/features/document/domain"
	documentusecase "apihorpug/internal/features/document/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *documentusecase.Service
}

type createDocumentRequest struct {
	DormitoryID  uuid.UUID                       `json:"dormitory_id"`
	TenantID     *uuid.UUID                      `json:"tenant_id"`
	RoomID       *uuid.UUID                      `json:"room_id"`
	Name         string                          `json:"name"`
	Category     documentdomain.DocumentCategory `json:"category"`
	FileURL      string                          `json:"file_url"`
	UploadedDate *time.Time                      `json:"uploaded_date"`
	Note         string                          `json:"note"`
}

type updateDocumentRequest struct {
	TenantID     *uuid.UUID                       `json:"tenant_id"`
	RoomID       *uuid.UUID                       `json:"room_id"`
	Name         *string                          `json:"name"`
	Category     *documentdomain.DocumentCategory `json:"category"`
	FileURL      *string                          `json:"file_url"`
	UploadedDate *time.Time                       `json:"uploaded_date"`
	Note         *string                          `json:"note"`
}

func NewHandler(usecase *documentusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

func parseUUIDQuery(c fiber.Ctx, name string) (*uuid.UUID, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, apierror.BadRequest("invalid " + name)
	}
	return &id, nil
}

func parseListFilters(c fiber.Ctx) (documentusecase.ListFilters, error) {
	dormitoryID, err := parseUUIDQuery(c, "dormitory_id")
	if err != nil {
		return documentusecase.ListFilters{}, err
	}
	tenantID, err := parseUUIDQuery(c, "tenant_id")
	if err != nil {
		return documentusecase.ListFilters{}, err
	}
	roomID, err := parseUUIDQuery(c, "room_id")
	if err != nil {
		return documentusecase.ListFilters{}, err
	}

	var category *documentdomain.DocumentCategory
	if raw := strings.TrimSpace(c.Query("category")); raw != "" {
		cat := documentdomain.DocumentCategory(raw)
		if !cat.Valid() {
			return documentusecase.ListFilters{}, apierror.BadRequest("invalid category")
		}
		category = &cat
	}

	return documentusecase.ListFilters{
		DormitoryID: dormitoryID,
		TenantID:    tenantID,
		RoomID:      roomID,
		Category:    category,
	}, nil
}

// List godoc
// @Summary List documents
// @Description Returns document records for roles with full dormitory access, otherwise only records under dormitories the caller manages. Optionally filter by dormitory, tenant, room or category.
// @Tags documents
// @Produce json
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param tenant_id query string false "Filter by tenant ID"
// @Param room_id query string false "Filter by room ID"
// @Param category query string false "Filter by category (contract, id_card, receipt, other)"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /documents [get]
func (h *Handler) List(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	filters, err := parseListFilters(c)
	if err != nil {
		return err
	}

	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	documents, total, err := h.usecase.List(ctx, requesterID, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list documents")
	}

	return apiresponse.Paginated(c, documents, page, perPage, total)
}

// Get godoc
// @Summary Get a document by ID
// @Tags documents
// @Produce json
// @Param id path string true "Document ID"
// @Success 200 {object} documentdomain.Document
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /documents/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid document id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	document, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, documentdomain.ErrDocumentNotFound) {
			return apierror.NotFound("document not found")
		}
		return apierror.Internal("failed to get document")
	}

	return apiresponse.OK(c, document)
}

// Create godoc
// @Summary Upload a document record
// @Description Records a general file attachment stored for a dormitory, optionally tied to the tenant or room it concerns.
// @Tags documents
// @Accept json
// @Produce json
// @Param request body createDocumentRequest true "Document payload"
// @Success 201 {object} documentdomain.Document
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /documents [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createDocumentRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	var uploadedDate time.Time
	if req.UploadedDate != nil {
		uploadedDate = *req.UploadedDate
	}

	document, err := h.usecase.Create(ctx, documentusecase.CreateInput{
		DormitoryID:  req.DormitoryID,
		TenantID:     req.TenantID,
		RoomID:       req.RoomID,
		Name:         req.Name,
		Category:     req.Category,
		FileURL:      req.FileURL,
		UploadedDate: uploadedDate,
		Note:         req.Note,
		CreatedBy:    &requesterID,
	})
	if err != nil {
		if errors.Is(err, documentdomain.ErrRequiredDocumentData) {
			return apierror.BadRequest("dormitory_id, name and file_url are required")
		}
		if errors.Is(err, documentdomain.ErrInvalidDocumentCategory) {
			return apierror.BadRequest("invalid document category")
		}
		if errors.Is(err, documentdomain.ErrDormitoryNotFound) {
			return apierror.NotFound("dormitory not found")
		}
		if errors.Is(err, documentdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		if errors.Is(err, documentdomain.ErrRoomNotFound) {
			return apierror.NotFound("room not found")
		}
		return apierror.Internal("failed to create document")
	}

	return apiresponse.Created(c, document)
}

// Update godoc
// @Summary Update a document record
// @Tags documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Param request body updateDocumentRequest true "Document payload"
// @Success 200 {object} documentdomain.Document
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /documents/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid document id")
	}

	var req updateDocumentRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	document, err := h.usecase.Update(ctx, id, requesterID, documentusecase.UpdateInput{
		TenantID:     req.TenantID,
		RoomID:       req.RoomID,
		Name:         req.Name,
		Category:     req.Category,
		FileURL:      req.FileURL,
		UploadedDate: req.UploadedDate,
		Note:         req.Note,
		UpdatedBy:    &requesterID,
	})
	if err != nil {
		if errors.Is(err, documentdomain.ErrDocumentNotFound) {
			return apierror.NotFound("document not found")
		}
		if errors.Is(err, documentdomain.ErrRequiredDocumentData) {
			return apierror.BadRequest("name and file_url must not be empty")
		}
		if errors.Is(err, documentdomain.ErrInvalidDocumentCategory) {
			return apierror.BadRequest("invalid document category")
		}
		if errors.Is(err, documentdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		if errors.Is(err, documentdomain.ErrRoomNotFound) {
			return apierror.NotFound("room not found")
		}
		return apierror.Internal("failed to update document")
	}

	return apiresponse.OK(c, document)
}

// Delete godoc
// @Summary Delete a document record
// @Tags documents
// @Produce json
// @Param id path string true "Document ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /documents/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid document id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID); err != nil {
		if errors.Is(err, documentdomain.ErrDocumentNotFound) {
			return apierror.NotFound("document not found")
		}
		return apierror.Internal("failed to delete document")
	}

	return apiresponse.Message(c, "document deleted")
}
