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

// parseColumnFilters reads the per-column filters, passed as f[<column>]=value.
// An unrecognised column is rejected rather than ignored: silently dropping it
// would return an unfiltered list that looks like a legitimate result.
func parseColumnFilters(c fiber.Ctx) (map[string]string, error) {
	columns := make(map[string]string)

	for key, value := range c.Queries() {
		if !strings.HasPrefix(key, "f[") || !strings.HasSuffix(key, "]") {
			continue
		}

		name := key[len("f[") : len(key)-len("]")]
		if _, ok := activitylogusecase.FilterColumns[name]; !ok {
			return nil, apierror.BadRequest("unsupported filter field: " + name)
		}

		if value = strings.TrimSpace(value); value != "" {
			columns[name] = value
		}
	}

	return columns, nil
}

// List godoc
// @Summary List activity logs
// @Tags activity-logs
// @Produce json
// @Param q query string false "Filter by user, action, entity type, description or IP"
// @Param f[column] query string false "Per-column substring filter, e.g. f[action]=create; column must be one of username, action, entity_type, description, ip_address"
// @Param user_id query string false "Filter by user ID"
// @Param entity_id query string false "Filter by entity ID"
// @Param date_from query string false "Filter by date, inclusive (YYYY-MM-DD)"
// @Param date_to query string false "Filter by date, inclusive (YYYY-MM-DD)"
// @Param sort query string false "Sort field: created_at, username, action, entity_type, description, ip_address (default created_at)"
// @Param order query string false "Sort direction: asc or desc"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /activity-logs [get]
func (h *Handler) List(c fiber.Ctx) error {
	filter := activitylogusecase.ListFilter{
		Search: strings.TrimSpace(c.Query("q")),
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

	columns, err := parseColumnFilters(c)
	if err != nil {
		return err
	}
	filter.Columns = columns

	sortKey := activitylogusecase.DefaultSortKey
	if raw := strings.TrimSpace(c.Query("sort")); raw != "" {
		if _, ok := activitylogusecase.SortColumns[raw]; !ok {
			return apierror.BadRequest("unsupported sort field")
		}
		sortKey = raw
	}
	filter.SortKey = sortKey

	switch strings.TrimSpace(c.Query("order")) {
	case "":
	case "asc":
		filter.SortDesc = false
	case "desc":
		filter.SortDesc = true
	default:
		return apierror.BadRequest("order must be asc or desc")
	}

	dateFrom, err := parseDateQuery(c, "date_from")
	if err != nil {
		return err
	}
	filter.DateFrom = dateFrom

	dateTo, err := parseDateQuery(c, "date_to")
	if err != nil {
		return err
	}
	if dateTo != nil {
		// created_at is a timestamp, so make the bound exclusive of the
		// following day to keep date_to inclusive of the whole day.
		endOfDay := dateTo.AddDate(0, 0, 1)
		filter.DateTo = &endOfDay
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
