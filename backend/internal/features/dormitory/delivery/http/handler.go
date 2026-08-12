package http

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	dormdomain "apihorpug/internal/features/dormitory/domain"
	dormusecase "apihorpug/internal/features/dormitory/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *dormusecase.Service
}

type createDormitoryRequest struct {
	Name        string      `json:"name"`
	Address     string      `json:"address"`
	Phone       string      `json:"phone"`
	Description string      `json:"description"`
	IsActive    *bool       `json:"is_active"`
	ManagerIDs  []uuid.UUID `json:"manager_ids"`
}

type updateDormitoryRequest struct {
	Name        *string      `json:"name"`
	Address     *string      `json:"address"`
	Phone       *string      `json:"phone"`
	Description *string      `json:"description"`
	IsActive    *bool        `json:"is_active"`
	ManagerIDs  *[]uuid.UUID `json:"manager_ids"`
}

func NewHandler(usecase *dormusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

const defaultActiveListLimit = 50

// List godoc
// @Summary List dormitories
// @Description Returns every dormitory for roles with full dormitory access, otherwise only the dormitories the caller manages.
// @Tags dormitories
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /dormitories [get]
func (h *Handler) List(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	dormitories, total, err := h.usecase.List(ctx, requesterID, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list dormitories")
	}

	return apiresponse.Paginated(c, dormitories, page, perPage, total)
}

// ListActive godoc
// @Summary List active dormitories
// @Description Returns active dormitories for roles with full dormitory access, otherwise only the active dormitories the caller manages. Intended for populating dormitory selectors; results are capped rather than paginated.
// @Tags dormitories
// @Produce json
// @Param q query string false "Filter by dormitory name"
// @Param limit query int false "Max results (default 50, max 100)"
// @Success 200 {array} dormdomain.Dormitory
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /dormitories/active [get]
func (h *Handler) ListActive(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	search := strings.TrimSpace(c.Query("q"))

	limit := defaultActiveListLimit
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return apierror.BadRequest("limit must be a positive integer")
		}
		limit = parsed
	}
	if limit > httputil.MaxPerPage {
		limit = httputil.MaxPerPage
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	dormitories, err := h.usecase.ListActive(ctx, requesterID, search, limit)
	if err != nil {
		return apierror.Internal("failed to list active dormitories")
	}

	return apiresponse.OK(c, dormitories)
}

// Get godoc
// @Summary Get a dormitory by ID
// @Tags dormitories
// @Produce json
// @Param id path string true "Dormitory ID"
// @Success 200 {object} dormdomain.Dormitory
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /dormitories/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid dormitory id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	dormitory, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, dormdomain.ErrDormitoryNotFound) {
			return apierror.NotFound("dormitory not found")
		}
		return apierror.Internal("failed to get dormitory")
	}

	return apiresponse.OK(c, dormitory)
}

// Create godoc
// @Summary Create a dormitory
// @Tags dormitories
// @Accept json
// @Produce json
// @Param request body createDormitoryRequest true "Dormitory payload"
// @Success 201 {object} dormdomain.Dormitory
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /dormitories [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createDormitoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	dormitory, err := h.usecase.Create(ctx, dormusecase.CreateInput{
		Name:        strings.TrimSpace(req.Name),
		Address:     req.Address,
		Phone:       req.Phone,
		Description: req.Description,
		IsActive:    isActive,
		ManagerIDs:  req.ManagerIDs,
		CreatedBy:   &requesterID,
	})
	if err != nil {
		if errors.Is(err, dormdomain.ErrRequiredDormitoryData) {
			return apierror.BadRequest("name is required")
		}
		if errors.Is(err, dormdomain.ErrManagerNotFound) {
			return apierror.BadRequest("one or more users not found")
		}
		return apierror.Internal("failed to create dormitory")
	}

	return apiresponse.Created(c, dormitory)
}

// Update godoc
// @Summary Update a dormitory
// @Tags dormitories
// @Accept json
// @Produce json
// @Param id path string true "Dormitory ID"
// @Param request body updateDormitoryRequest true "Dormitory payload"
// @Success 200 {object} dormdomain.Dormitory
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /dormitories/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid dormitory id")
	}

	var req updateDormitoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	dormitory, err := h.usecase.Update(ctx, id, requesterID, dormusecase.UpdateInput{
		Name:        req.Name,
		Address:     req.Address,
		Phone:       req.Phone,
		Description: req.Description,
		IsActive:    req.IsActive,
		ManagerIDs:  req.ManagerIDs,
		UpdatedBy:   &requesterID,
	})
	if err != nil {
		if errors.Is(err, dormdomain.ErrDormitoryNotFound) {
			return apierror.NotFound("dormitory not found")
		}
		if errors.Is(err, dormdomain.ErrRequiredDormitoryData) {
			return apierror.BadRequest("name cannot be empty")
		}
		if errors.Is(err, dormdomain.ErrManagerNotFound) {
			return apierror.BadRequest("one or more users not found")
		}
		return apierror.Internal("failed to update dormitory")
	}

	return apiresponse.OK(c, dormitory)
}

// Delete godoc
// @Summary Delete a dormitory
// @Tags dormitories
// @Produce json
// @Param id path string true "Dormitory ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /dormitories/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid dormitory id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID); err != nil {
		if errors.Is(err, dormdomain.ErrDormitoryNotFound) {
			return apierror.NotFound("dormitory not found")
		}
		return apierror.Internal("failed to delete dormitory")
	}

	return apiresponse.Message(c, "dormitory deleted")
}
