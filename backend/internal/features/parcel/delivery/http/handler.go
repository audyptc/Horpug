package http

import (
	"context"
	"errors"
	"strings"
	"time"

	parceldomain "apihorpug/internal/features/parcel/domain"
	parcelusecase "apihorpug/internal/features/parcel/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *parcelusecase.Service
}

type createParcelRequest struct {
	TenantID       uuid.UUID                 `json:"tenant_id"`
	RoomID         *uuid.UUID                `json:"room_id"`
	Courier        string                    `json:"courier"`
	TrackingNumber string                    `json:"tracking_number"`
	Status         parceldomain.ParcelStatus `json:"status"`
	ReceivedDate   time.Time                 `json:"received_date"`
	Note           string                    `json:"note"`
}

type updateParcelRequest struct {
	RoomID         *uuid.UUID                 `json:"room_id"`
	Courier        *string                    `json:"courier"`
	TrackingNumber *string                    `json:"tracking_number"`
	Status         *parceldomain.ParcelStatus `json:"status"`
	ReceivedDate   *time.Time                 `json:"received_date"`
	Note           *string                    `json:"note"`
}

func NewHandler(usecase *parcelusecase.Service) *Handler {
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

func parseListFilters(c fiber.Ctx) (parcelusecase.ListFilters, error) {
	tenantID, err := parseUUIDQuery(c, "tenant_id")
	if err != nil {
		return parcelusecase.ListFilters{}, err
	}
	roomID, err := parseUUIDQuery(c, "room_id")
	if err != nil {
		return parcelusecase.ListFilters{}, err
	}
	dormitoryID, err := parseUUIDQuery(c, "dormitory_id")
	if err != nil {
		return parcelusecase.ListFilters{}, err
	}

	var status *parceldomain.ParcelStatus
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		st := parceldomain.ParcelStatus(raw)
		if !st.Valid() {
			return parcelusecase.ListFilters{}, apierror.BadRequest("invalid status")
		}
		status = &st
	}

	return parcelusecase.ListFilters{
		TenantID:    tenantID,
		RoomID:      roomID,
		DormitoryID: dormitoryID,
		Status:      status,
	}, nil
}

// List godoc
// @Summary List parcels
// @Description Returns parcel records for roles with full dormitory access, otherwise only records whose room belongs to a dormitory the caller manages (records without a room are only visible to roles with full access). Optionally filter by tenant, room, dormitory or status.
// @Tags parcels
// @Produce json
// @Param tenant_id query string false "Filter by tenant ID"
// @Param room_id query string false "Filter by room ID"
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param status query string false "Filter by status (pending, picked_up, returned)"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /parcels [get]
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

	parcels, total, err := h.usecase.List(ctx, requesterID, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list parcels")
	}

	return apiresponse.Paginated(c, parcels, page, perPage, total)
}

// Get godoc
// @Summary Get a parcel by ID
// @Tags parcels
// @Produce json
// @Param id path string true "Parcel ID"
// @Success 200 {object} parceldomain.Parcel
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /parcels/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid parcel id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	parcel, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, parceldomain.ErrParcelNotFound) {
			return apierror.NotFound("parcel not found")
		}
		return apierror.Internal("failed to get parcel")
	}

	return apiresponse.OK(c, parcel)
}

// Create godoc
// @Summary Record a parcel received for a tenant
// @Description Records a package delivered to the dormitory office for a tenant, optionally tied to the room they occupy so it can be scoped to a dormitory.
// @Tags parcels
// @Accept json
// @Produce json
// @Param request body createParcelRequest true "Parcel payload"
// @Success 201 {object} parceldomain.Parcel
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /parcels [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createParcelRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	parcel, err := h.usecase.Create(ctx, parcelusecase.CreateInput{
		TenantID:       req.TenantID,
		RoomID:         req.RoomID,
		Courier:        req.Courier,
		TrackingNumber: req.TrackingNumber,
		Status:         req.Status,
		ReceivedDate:   req.ReceivedDate,
		Note:           req.Note,
		CreatedBy:      &requesterID,
	})
	if err != nil {
		if errors.Is(err, parceldomain.ErrRequiredParcelData) {
			return apierror.BadRequest("tenant_id and received_date are required")
		}
		if errors.Is(err, parceldomain.ErrInvalidParcelStatus) {
			return apierror.BadRequest("invalid parcel status")
		}
		if errors.Is(err, parceldomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		if errors.Is(err, parceldomain.ErrRoomNotFound) {
			return apierror.NotFound("room not found")
		}
		return apierror.Internal("failed to record parcel")
	}

	return apiresponse.Created(c, parcel)
}

// Update godoc
// @Summary Update a parcel record
// @Tags parcels
// @Accept json
// @Produce json
// @Param id path string true "Parcel ID"
// @Param request body updateParcelRequest true "Parcel payload"
// @Success 200 {object} parceldomain.Parcel
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /parcels/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid parcel id")
	}

	var req updateParcelRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	parcel, err := h.usecase.Update(ctx, id, requesterID, parcelusecase.UpdateInput{
		RoomID:         req.RoomID,
		Courier:        req.Courier,
		TrackingNumber: req.TrackingNumber,
		Status:         req.Status,
		ReceivedDate:   req.ReceivedDate,
		Note:           req.Note,
		UpdatedBy:      &requesterID,
	})
	if err != nil {
		if errors.Is(err, parceldomain.ErrParcelNotFound) {
			return apierror.NotFound("parcel not found")
		}
		if errors.Is(err, parceldomain.ErrInvalidParcelStatus) {
			return apierror.BadRequest("invalid parcel status")
		}
		if errors.Is(err, parceldomain.ErrRoomNotFound) {
			return apierror.NotFound("room not found")
		}
		return apierror.Internal("failed to update parcel")
	}

	return apiresponse.OK(c, parcel)
}

// Delete godoc
// @Summary Delete a parcel record
// @Tags parcels
// @Produce json
// @Param id path string true "Parcel ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /parcels/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid parcel id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID); err != nil {
		if errors.Is(err, parceldomain.ErrParcelNotFound) {
			return apierror.NotFound("parcel not found")
		}
		return apierror.Internal("failed to delete parcel")
	}

	return apiresponse.Message(c, "parcel deleted")
}
