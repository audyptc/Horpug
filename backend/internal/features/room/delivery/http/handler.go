package http

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	roomdomain "apihorpug/internal/features/room/domain"
	roomusecase "apihorpug/internal/features/room/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *roomusecase.Service
}

type createRoomRequest struct {
	DormitoryID uuid.UUID             `json:"dormitory_id"`
	RoomTypeID  uuid.UUID             `json:"room_type_id"`
	RoomNumber  string                `json:"room_number"`
	Floor       int                   `json:"floor"`
	Status      roomdomain.RoomStatus `json:"status"`
	IsActive    *bool                 `json:"is_active"`
}

type updateRoomRequest struct {
	RoomTypeID *uuid.UUID             `json:"room_type_id"`
	RoomNumber *string                `json:"room_number"`
	Floor      *int                   `json:"floor"`
	Status     *roomdomain.RoomStatus `json:"status"`
	IsActive   *bool                  `json:"is_active"`
}

func NewHandler(usecase *roomusecase.Service) *Handler {
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
// @Summary List rooms
// @Description Returns every room for roles with full dormitory access, otherwise only rooms under dormitories the caller manages. Optionally filter by dormitory.
// @Tags rooms
// @Produce json
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /rooms [get]
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

	rooms, total, err := h.usecase.List(ctx, requesterID, dormitoryID, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list rooms")
	}

	return apiresponse.Paginated(c, rooms, page, perPage, total)
}

// ListActive godoc
// @Summary List active rooms
// @Description Returns active rooms for roles with full dormitory access, otherwise only the active rooms under dormitories the caller manages. Intended for populating room selectors; results are capped rather than paginated.
// @Tags rooms
// @Produce json
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param q query string false "Filter by room number"
// @Param limit query int false "Max results (default 50, max 100)"
// @Success 200 {array} roomdomain.Room
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /rooms/active [get]
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

	rooms, err := h.usecase.ListActive(ctx, requesterID, dormitoryID, search, limit)
	if err != nil {
		return apierror.Internal("failed to list active rooms")
	}

	return apiresponse.OK(c, rooms)
}

// Get godoc
// @Summary Get a room by ID
// @Tags rooms
// @Produce json
// @Param id path string true "Room ID"
// @Success 200 {object} roomdomain.Room
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /rooms/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid room id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	room, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, roomdomain.ErrRoomNotFound) {
			return apierror.NotFound("room not found")
		}
		return apierror.Internal("failed to get room")
	}

	return apiresponse.OK(c, room)
}

// Create godoc
// @Summary Create a room
// @Tags rooms
// @Accept json
// @Produce json
// @Param request body createRoomRequest true "Room payload"
// @Success 201 {object} roomdomain.Room
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /rooms [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createRoomRequest
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

	room, err := h.usecase.Create(ctx, roomusecase.CreateInput{
		DormitoryID: req.DormitoryID,
		RoomTypeID:  req.RoomTypeID,
		RoomNumber:  strings.TrimSpace(req.RoomNumber),
		Floor:       req.Floor,
		Status:      req.Status,
		IsActive:    isActive,
		CreatedBy:   &requesterID,
	})
	if err != nil {
		if errors.Is(err, roomdomain.ErrRequiredRoomData) {
			return apierror.BadRequest("room_number, dormitory_id and room_type_id are required")
		}
		if errors.Is(err, roomdomain.ErrInvalidRoomStatus) {
			return apierror.BadRequest("invalid room status")
		}
		if errors.Is(err, roomdomain.ErrDormitoryNotFound) {
			return apierror.NotFound("dormitory not found")
		}
		if errors.Is(err, roomdomain.ErrRoomTypeNotFound) {
			return apierror.NotFound("room type not found in this dormitory")
		}
		if errors.Is(err, roomdomain.ErrRoomNumberExists) {
			return apierror.Conflict("room number already exists in this dormitory")
		}
		return apierror.Internal("failed to create room")
	}

	return apiresponse.Created(c, room)
}

// Update godoc
// @Summary Update a room
// @Tags rooms
// @Accept json
// @Produce json
// @Param id path string true "Room ID"
// @Param request body updateRoomRequest true "Room payload"
// @Success 200 {object} roomdomain.Room
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /rooms/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid room id")
	}

	var req updateRoomRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	room, err := h.usecase.Update(ctx, id, requesterID, roomusecase.UpdateInput{
		RoomTypeID: req.RoomTypeID,
		RoomNumber: req.RoomNumber,
		Floor:      req.Floor,
		Status:     req.Status,
		IsActive:   req.IsActive,
		UpdatedBy:  &requesterID,
	})
	if err != nil {
		if errors.Is(err, roomdomain.ErrRoomNotFound) {
			return apierror.NotFound("room not found")
		}
		if errors.Is(err, roomdomain.ErrRequiredRoomData) {
			return apierror.BadRequest("room_number cannot be empty")
		}
		if errors.Is(err, roomdomain.ErrInvalidRoomStatus) {
			return apierror.BadRequest("invalid room status")
		}
		if errors.Is(err, roomdomain.ErrRoomTypeNotFound) {
			return apierror.NotFound("room type not found in this dormitory")
		}
		if errors.Is(err, roomdomain.ErrRoomNumberExists) {
			return apierror.Conflict("room number already exists in this dormitory")
		}
		return apierror.Internal("failed to update room")
	}

	return apiresponse.OK(c, room)
}

// Delete godoc
// @Summary Delete a room
// @Tags rooms
// @Produce json
// @Param id path string true "Room ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /rooms/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid room id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID); err != nil {
		if errors.Is(err, roomdomain.ErrRoomNotFound) {
			return apierror.NotFound("room not found")
		}
		return apierror.Internal("failed to delete room")
	}

	return apiresponse.Message(c, "room deleted")
}
