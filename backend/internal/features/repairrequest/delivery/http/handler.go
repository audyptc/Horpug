package http

import (
	"context"
	"errors"
	"strings"
	"time"

	repairrequestdomain "apihorpug/internal/features/repairrequest/domain"
	repairrequestusecase "apihorpug/internal/features/repairrequest/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *repairrequestusecase.Service
}

type createRepairRequestRequest struct {
	RoomID       uuid.UUID                          `json:"room_id"`
	TenantID     *uuid.UUID                         `json:"tenant_id"`
	Category     repairrequestdomain.RepairCategory `json:"category"`
	Description  string                             `json:"description"`
	Status       repairrequestdomain.RepairStatus   `json:"status"`
	ReportedDate time.Time                          `json:"reported_date"`
}

type updateRepairRequestRequest struct {
	TenantID     *uuid.UUID                          `json:"tenant_id"`
	Category     *repairrequestdomain.RepairCategory `json:"category"`
	Description  *string                             `json:"description"`
	Status       *repairrequestdomain.RepairStatus   `json:"status"`
	ReportedDate *time.Time                          `json:"reported_date"`
}

func NewHandler(usecase *repairrequestusecase.Service) *Handler {
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
		if _, ok := repairrequestusecase.FilterColumns[name]; !ok {
			return nil, apierror.BadRequest("unsupported filter field: " + name)
		}

		if value = strings.TrimSpace(value); value != "" {
			columns[name] = value
		}
	}

	return columns, nil
}

func parseListFilters(c fiber.Ctx) (repairrequestusecase.ListFilters, error) {
	roomID, err := parseUUIDQuery(c, "room_id")
	if err != nil {
		return repairrequestusecase.ListFilters{}, err
	}
	dormitoryID, err := parseUUIDQuery(c, "dormitory_id")
	if err != nil {
		return repairrequestusecase.ListFilters{}, err
	}
	tenantID, err := parseUUIDQuery(c, "tenant_id")
	if err != nil {
		return repairrequestusecase.ListFilters{}, err
	}
	columns, err := parseColumnFilters(c)
	if err != nil {
		return repairrequestusecase.ListFilters{}, err
	}

	var category *repairrequestdomain.RepairCategory
	if raw := strings.TrimSpace(c.Query("category")); raw != "" {
		cat := repairrequestdomain.RepairCategory(raw)
		if !cat.Valid() {
			return repairrequestusecase.ListFilters{}, apierror.BadRequest("invalid category")
		}
		category = &cat
	}

	var status *repairrequestdomain.RepairStatus
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		st := repairrequestdomain.RepairStatus(raw)
		if !st.Valid() {
			return repairrequestusecase.ListFilters{}, apierror.BadRequest("invalid status")
		}
		status = &st
	}

	sortKey := repairrequestusecase.DefaultSortKey
	if raw := strings.TrimSpace(c.Query("sort")); raw != "" {
		if _, ok := repairrequestusecase.SortColumns[raw]; !ok {
			return repairrequestusecase.ListFilters{}, apierror.BadRequest("unsupported sort field")
		}
		sortKey = raw
	}

	// Newest-first is the useful default for an unsorted listing, but an
	// explicit sort field reads more naturally ascending.
	sortDesc := sortKey == repairrequestusecase.DefaultSortKey
	switch strings.TrimSpace(c.Query("order")) {
	case "":
	case "asc":
		sortDesc = false
	case "desc":
		sortDesc = true
	default:
		return repairrequestusecase.ListFilters{}, apierror.BadRequest("order must be asc or desc")
	}

	return repairrequestusecase.ListFilters{
		RoomID:      roomID,
		DormitoryID: dormitoryID,
		TenantID:    tenantID,
		Category:    category,
		Status:      status,
		Search:      strings.TrimSpace(c.Query("q")),
		Columns:     columns,
		SortKey:     sortKey,
		SortDesc:    sortDesc,
	}, nil
}

// List godoc
// @Summary List repair requests
// @Description Returns every repair request for roles with full dormitory access, otherwise only requests under dormitories the caller manages. Optionally filter by room, dormitory, tenant, category or status.
// @Tags repair-requests
// @Produce json
// @Param room_id query string false "Filter by room ID"
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param tenant_id query string false "Filter by reporting tenant ID"
// @Param category query string false "Filter by category (electrical, plumbing, furniture, aircon, other)"
// @Param status query string false "Filter by status (pending, in_progress, completed, cancelled)"
// @Param q query string false "Filter by room number, dormitory name, tenant name or description"
// @Param f[column] query string false "Per-column substring filter, e.g. f[room_number]=101; column must be one of room_number, dormitory_name, tenant_name, description"
// @Param sort query string false "Sort field: room_number, tenant_name, category, status, reported_date, description (default reported_date)"
// @Param order query string false "Sort direction: asc or desc"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /repair-requests [get]
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

	requests, total, err := h.usecase.List(ctx, requesterID, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list repair requests")
	}

	return apiresponse.Paginated(c, requests, page, perPage, total)
}

// Get godoc
// @Summary Get a repair request by ID
// @Tags repair-requests
// @Produce json
// @Param id path string true "Repair request ID"
// @Success 200 {object} repairrequestdomain.RepairRequest
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /repair-requests/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid repair request id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	request, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, repairrequestdomain.ErrRepairRequestNotFound) {
			return apierror.NotFound("repair request not found")
		}
		return apierror.Internal("failed to get repair request")
	}

	return apiresponse.OK(c, request)
}

// Create godoc
// @Summary Report a repair request
// @Description Records a maintenance issue reported for a room, optionally attributed to the tenant who reported it.
// @Tags repair-requests
// @Accept json
// @Produce json
// @Param request body createRepairRequestRequest true "Repair request payload"
// @Success 201 {object} repairrequestdomain.RepairRequest
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /repair-requests [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createRepairRequestRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	request, err := h.usecase.Create(ctx, repairrequestusecase.CreateInput{
		RoomID:       req.RoomID,
		TenantID:     req.TenantID,
		Category:     req.Category,
		Description:  req.Description,
		Status:       req.Status,
		ReportedDate: req.ReportedDate,
		CreatedBy:    &requesterID,
	})
	if err != nil {
		if errors.Is(err, repairrequestdomain.ErrRequiredRepairRequestData) {
			return apierror.BadRequest("room_id, description and reported_date are required")
		}
		if errors.Is(err, repairrequestdomain.ErrInvalidCategory) {
			return apierror.BadRequest("invalid repair category")
		}
		if errors.Is(err, repairrequestdomain.ErrInvalidStatus) {
			return apierror.BadRequest("invalid repair status")
		}
		if errors.Is(err, repairrequestdomain.ErrRoomNotFound) {
			return apierror.NotFound("room not found")
		}
		if errors.Is(err, repairrequestdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		return apierror.Internal("failed to create repair request")
	}

	return apiresponse.Created(c, request)
}

// Update godoc
// @Summary Update a repair request
// @Tags repair-requests
// @Accept json
// @Produce json
// @Param id path string true "Repair request ID"
// @Param request body updateRepairRequestRequest true "Repair request payload"
// @Success 200 {object} repairrequestdomain.RepairRequest
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /repair-requests/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid repair request id")
	}

	var req updateRepairRequestRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	request, err := h.usecase.Update(ctx, id, requesterID, repairrequestusecase.UpdateInput{
		TenantID:     req.TenantID,
		Category:     req.Category,
		Description:  req.Description,
		Status:       req.Status,
		ReportedDate: req.ReportedDate,
		UpdatedBy:    &requesterID,
	})
	if err != nil {
		if errors.Is(err, repairrequestdomain.ErrRepairRequestNotFound) {
			return apierror.NotFound("repair request not found")
		}
		if errors.Is(err, repairrequestdomain.ErrRequiredRepairRequestData) {
			return apierror.BadRequest("description must not be empty")
		}
		if errors.Is(err, repairrequestdomain.ErrInvalidCategory) {
			return apierror.BadRequest("invalid repair category")
		}
		if errors.Is(err, repairrequestdomain.ErrInvalidStatus) {
			return apierror.BadRequest("invalid repair status")
		}
		if errors.Is(err, repairrequestdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		return apierror.Internal("failed to update repair request")
	}

	return apiresponse.OK(c, request)
}

// Delete godoc
// @Summary Delete a repair request
// @Tags repair-requests
// @Produce json
// @Param id path string true "Repair request ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /repair-requests/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid repair request id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID); err != nil {
		if errors.Is(err, repairrequestdomain.ErrRepairRequestNotFound) {
			return apierror.NotFound("repair request not found")
		}
		return apierror.Internal("failed to delete repair request")
	}

	return apiresponse.Message(c, "repair request deleted")
}
