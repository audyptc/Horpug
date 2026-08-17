package http

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	roomtypedomain "apihorpug/internal/features/roomtype/domain"
	roomtypeusecase "apihorpug/internal/features/roomtype/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *roomtypeusecase.Service
}

type createRoomTypeRequest struct {
	DormitoryID uuid.UUID `json:"dormitory_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	IsActive    *bool     `json:"is_active"`
}

type updateRoomTypeRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
}

func NewHandler(usecase *roomtypeusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

const defaultActiveListLimit = 50

func parseDormitoryIDQuery(c fiber.Ctx) (*uuid.UUID, error) {
	raw := strings.TrimSpace(c.Query("dormitory_id"))
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, apierror.BadRequest("invalid dormitory_id")
	}
	return &id, nil
}

// List godoc
// @Summary List room types
// @Description Returns every room type for roles with full dormitory access, otherwise only room types under dormitories the caller manages. Optionally filter by dormitory.
// @Tags room-types
// @Produce json
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /room-types [get]
func (h *Handler) List(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	dormitoryID, err := parseDormitoryIDQuery(c)
	if err != nil {
		return err
	}

	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	roomTypes, total, err := h.usecase.List(ctx, requesterID, dormitoryID, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list room types")
	}

	return apiresponse.Paginated(c, roomTypes, page, perPage, total)
}

// ListActive godoc
// @Summary List active room types
// @Description Returns active room types for roles with full dormitory access, otherwise only the active room types under dormitories the caller manages. Intended for populating room type selectors; results are capped rather than paginated.
// @Tags room-types
// @Produce json
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param q query string false "Filter by room type name"
// @Param limit query int false "Max results (default 50, max 100)"
// @Success 200 {array} roomtypedomain.RoomType
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /room-types/active [get]
func (h *Handler) ListActive(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	dormitoryID, err := parseDormitoryIDQuery(c)
	if err != nil {
		return err
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

	roomTypes, err := h.usecase.ListActive(ctx, requesterID, dormitoryID, search, limit)
	if err != nil {
		return apierror.Internal("failed to list active room types")
	}

	return apiresponse.OK(c, roomTypes)
}

// Get godoc
// @Summary Get a room type by ID
// @Tags room-types
// @Produce json
// @Param id path string true "Room type ID"
// @Success 200 {object} roomtypedomain.RoomType
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /room-types/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid room type id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	roomType, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, roomtypedomain.ErrRoomTypeNotFound) {
			return apierror.NotFound("room type not found")
		}
		return apierror.Internal("failed to get room type")
	}

	return apiresponse.OK(c, roomType)
}

// CheckDeletion godoc
// @Summary Check whether a room type can be deleted
// @Tags room-types
// @Produce json
// @Param id path string true "Room type ID"
// @Success 200 {object} roomtypeusecase.DeletionCheck
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /room-types/{id}/deletion-check [get]
func (h *Handler) CheckDeletion(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid room type id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	check, err := h.usecase.CheckDeletion(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, roomtypedomain.ErrRoomTypeNotFound) {
			return apierror.NotFound("room type not found")
		}
		return apierror.Internal("failed to check room type deletion")
	}

	return apiresponse.OK(c, check)
}

// Create godoc
// @Summary Create a room type
// @Tags room-types
// @Accept json
// @Produce json
// @Param request body createRoomTypeRequest true "Room type payload"
// @Success 201 {object} roomtypedomain.RoomType
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /room-types [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createRoomTypeRequest
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

	roomType, err := h.usecase.Create(ctx, roomtypeusecase.CreateInput{
		DormitoryID: req.DormitoryID,
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Price:       req.Price,
		IsActive:    isActive,
		CreatedBy:   &requesterID,
	})
	if err != nil {
		if errors.Is(err, roomtypedomain.ErrRequiredRoomTypeData) {
			return apierror.BadRequest("name and dormitory_id are required")
		}
		if errors.Is(err, roomtypedomain.ErrInvalidRoomTypePrice) {
			return apierror.BadRequest("price must not be negative")
		}
		if errors.Is(err, roomtypedomain.ErrDormitoryNotFound) {
			return apierror.NotFound("dormitory not found")
		}
		if errors.Is(err, roomtypedomain.ErrRoomTypeNameExists) {
			return apierror.Conflict("room type name already exists in this dormitory")
		}
		return apierror.Internal("failed to create room type")
	}

	return apiresponse.Created(c, roomType)
}

// Update godoc
// @Summary Update a room type
// @Tags room-types
// @Accept json
// @Produce json
// @Param id path string true "Room type ID"
// @Param request body updateRoomTypeRequest true "Room type payload"
// @Success 200 {object} roomtypedomain.RoomType
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /room-types/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid room type id")
	}

	var req updateRoomTypeRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	roomType, err := h.usecase.Update(ctx, id, requesterID, roomtypeusecase.UpdateInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		IsActive:    req.IsActive,
		UpdatedBy:   &requesterID,
	})
	if err != nil {
		if errors.Is(err, roomtypedomain.ErrRoomTypeNotFound) {
			return apierror.NotFound("room type not found")
		}
		if errors.Is(err, roomtypedomain.ErrRequiredRoomTypeData) {
			return apierror.BadRequest("name cannot be empty")
		}
		if errors.Is(err, roomtypedomain.ErrInvalidRoomTypePrice) {
			return apierror.BadRequest("price must not be negative")
		}
		if errors.Is(err, roomtypedomain.ErrRoomTypeNameExists) {
			return apierror.Conflict("room type name already exists in this dormitory")
		}
		return apierror.Internal("failed to update room type")
	}

	return apiresponse.OK(c, roomType)
}

// Delete godoc
// @Summary Delete a room type
// @Tags room-types
// @Produce json
// @Param id path string true "Room type ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /room-types/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid room type id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID); err != nil {
		if errors.Is(err, roomtypedomain.ErrRoomTypeNotFound) {
			return apierror.NotFound("room type not found")
		}
		if errors.Is(err, roomtypedomain.ErrRoomTypeHasRooms) {
			return apierror.Conflict("room type has rooms and cannot be deleted")
		}
		return apierror.Internal("failed to delete room type")
	}

	return apiresponse.Message(c, "room type deleted")
}
