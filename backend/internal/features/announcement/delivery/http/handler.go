package http

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	announcementdomain "apihorpug/internal/features/announcement/domain"
	announcementusecase "apihorpug/internal/features/announcement/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *announcementusecase.Service
}

type createAnnouncementRequest struct {
	DormitoryID   uuid.UUID  `json:"dormitory_id"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	IsPublished   *bool      `json:"is_published"`
	PublishedDate *time.Time `json:"published_date"`
}

type updateAnnouncementRequest struct {
	Title         *string    `json:"title"`
	Content       *string    `json:"content"`
	IsPublished   *bool      `json:"is_published"`
	PublishedDate *time.Time `json:"published_date"`
}

func NewHandler(usecase *announcementusecase.Service) *Handler {
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

func parseDateQuery(c fiber.Ctx, name string) (*time.Time, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	value, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, apierror.BadRequest("invalid " + name)
	}
	return &value, nil
}

func parseBoolQuery(c fiber.Ctx, name string) (*bool, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, apierror.BadRequest("invalid " + name)
	}
	return &value, nil
}

func parseListFilters(c fiber.Ctx) (announcementusecase.ListFilters, error) {
	dormitoryID, err := parseUUIDQuery(c, "dormitory_id")
	if err != nil {
		return announcementusecase.ListFilters{}, err
	}
	isPublished, err := parseBoolQuery(c, "is_published")
	if err != nil {
		return announcementusecase.ListFilters{}, err
	}
	dateFrom, err := parseDateQuery(c, "date_from")
	if err != nil {
		return announcementusecase.ListFilters{}, err
	}
	dateTo, err := parseDateQuery(c, "date_to")
	if err != nil {
		return announcementusecase.ListFilters{}, err
	}

	return announcementusecase.ListFilters{
		DormitoryID: dormitoryID,
		IsPublished: isPublished,
		DateFrom:    dateFrom,
		DateTo:      dateTo,
	}, nil
}

// List godoc
// @Summary List announcements
// @Description Returns every announcement for roles with full dormitory access, otherwise only announcements under dormitories the caller manages. Optionally filter by dormitory, published status or published date range.
// @Tags announcements
// @Produce json
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param is_published query bool false "Filter by published status"
// @Param date_from query string false "Filter by published date, inclusive (YYYY-MM-DD)"
// @Param date_to query string false "Filter by published date, inclusive (YYYY-MM-DD)"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /announcements [get]
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

	announcements, total, err := h.usecase.List(ctx, requesterID, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list announcements")
	}

	return apiresponse.Paginated(c, announcements, page, perPage, total)
}

// Get godoc
// @Summary Get an announcement by ID
// @Tags announcements
// @Produce json
// @Param id path string true "Announcement ID"
// @Success 200 {object} announcementdomain.Announcement
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /announcements/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid announcement id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	announcement, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, announcementdomain.ErrAnnouncementNotFound) {
			return apierror.NotFound("announcement not found")
		}
		return apierror.Internal("failed to get announcement")
	}

	return apiresponse.OK(c, announcement)
}

// Create godoc
// @Summary Post an announcement
// @Description Posts a notice (e.g. maintenance schedule, rule change, event) to tenants of a dormitory.
// @Tags announcements
// @Accept json
// @Produce json
// @Param request body createAnnouncementRequest true "Announcement payload"
// @Success 201 {object} announcementdomain.Announcement
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /announcements [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createAnnouncementRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	var publishedDate time.Time
	if req.PublishedDate != nil {
		publishedDate = *req.PublishedDate
	}

	announcement, err := h.usecase.Create(ctx, announcementusecase.CreateInput{
		DormitoryID:   req.DormitoryID,
		Title:         req.Title,
		Content:       req.Content,
		IsPublished:   req.IsPublished,
		PublishedDate: publishedDate,
		CreatedBy:     &requesterID,
	})
	if err != nil {
		if errors.Is(err, announcementdomain.ErrRequiredAnnouncementData) {
			return apierror.BadRequest("dormitory_id and title are required")
		}
		if errors.Is(err, announcementdomain.ErrDormitoryNotFound) {
			return apierror.NotFound("dormitory not found")
		}
		return apierror.Internal("failed to create announcement")
	}

	return apiresponse.Created(c, announcement)
}

// Update godoc
// @Summary Update an announcement
// @Tags announcements
// @Accept json
// @Produce json
// @Param id path string true "Announcement ID"
// @Param request body updateAnnouncementRequest true "Announcement payload"
// @Success 200 {object} announcementdomain.Announcement
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /announcements/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid announcement id")
	}

	var req updateAnnouncementRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	announcement, err := h.usecase.Update(ctx, id, requesterID, announcementusecase.UpdateInput{
		Title:         req.Title,
		Content:       req.Content,
		IsPublished:   req.IsPublished,
		PublishedDate: req.PublishedDate,
		UpdatedBy:     &requesterID,
	})
	if err != nil {
		if errors.Is(err, announcementdomain.ErrAnnouncementNotFound) {
			return apierror.NotFound("announcement not found")
		}
		if errors.Is(err, announcementdomain.ErrRequiredAnnouncementData) {
			return apierror.BadRequest("title must not be empty")
		}
		return apierror.Internal("failed to update announcement")
	}

	return apiresponse.OK(c, announcement)
}

// Delete godoc
// @Summary Delete an announcement
// @Tags announcements
// @Produce json
// @Param id path string true "Announcement ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /announcements/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid announcement id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID); err != nil {
		if errors.Is(err, announcementdomain.ErrAnnouncementNotFound) {
			return apierror.NotFound("announcement not found")
		}
		return apierror.Internal("failed to delete announcement")
	}

	return apiresponse.Message(c, "announcement deleted")
}
