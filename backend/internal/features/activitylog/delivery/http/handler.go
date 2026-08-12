package http

import (
	"context"
	"errors"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *activitylogusecase.Service
}

type createActivityLogRequest struct {
	UserID      *uuid.UUID `json:"user_id"`
	Action      string     `json:"action"`
	EntityType  string     `json:"entity_type"`
	EntityID    *uuid.UUID `json:"entity_id"`
	Description string     `json:"description"`
}

func NewHandler(usecase *activitylogusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

// List godoc
// @Summary List activity logs
// @Tags activity-logs
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param entity_type query string false "Filter by entity type"
// @Param entity_id query string false "Filter by entity ID"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /activity-logs [get]
func (h *Handler) List(c fiber.Ctx) error {
	filter := activitylogusecase.ListFilter{
		EntityType: strings.TrimSpace(c.Query("entity_type")),
	}

	if raw := c.Query("user_id"); raw != "" {
		userID, err := uuid.Parse(raw)
		if err != nil {
			return apierror.BadRequest("invalid user_id")
		}
		filter.UserID = &userID
	}

	if raw := c.Query("entity_id"); raw != "" {
		entityID, err := uuid.Parse(raw)
		if err != nil {
			return apierror.BadRequest("invalid entity_id")
		}
		filter.EntityID = &entityID
	}

	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	filter.Limit = perPage
	filter.Offset = offset

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	logs, total, err := h.usecase.List(ctx, filter)
	if err != nil {
		return apierror.Internal("failed to list activity logs")
	}

	return apiresponse.Paginated(c, logs, page, perPage, total)
}

// Get godoc
// @Summary Get an activity log by ID
// @Tags activity-logs
// @Produce json
// @Param id path string true "Activity Log ID"
// @Success 200 {object} activitylogdomain.ActivityLog
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /activity-logs/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid activity log id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	log, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, activitylogdomain.ErrActivityLogNotFound) {
			return apierror.NotFound("activity log not found")
		}
		return apierror.Internal("failed to get activity log")
	}

	return apiresponse.OK(c, log)
}

// Create godoc
// @Summary Create an activity log entry
// @Tags activity-logs
// @Accept json
// @Produce json
// @Param request body createActivityLogRequest true "Activity log payload"
// @Success 201 {object} activitylogdomain.ActivityLog
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /activity-logs [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createActivityLogRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	req.Action = strings.TrimSpace(req.Action)
	if req.Action == "" {
		return apierror.BadRequest("action is required")
	}

	req.EntityType = strings.TrimSpace(req.EntityType)
	if req.EntityType == "" {
		return apierror.BadRequest("entity_type is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	createdLog, err := h.usecase.Create(ctx, activitylogusecase.CreateInput{
		UserID:      req.UserID,
		Action:      req.Action,
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		Description: req.Description,
		IPAddress:   c.IP(),
	})
	if err != nil {
		return apierror.Internal("failed to create activity log")
	}

	return apiresponse.Created(c, createdLog)
}
