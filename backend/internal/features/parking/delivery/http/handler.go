package http

import (
	"context"
	"errors"
	"strings"
	"time"

	parkingdomain "apihorpug/internal/features/parking/domain"
	parkingusecase "apihorpug/internal/features/parking/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *parkingusecase.Service
}

type createParkingRequest struct {
	TenantID     uuid.UUID                 `json:"tenant_id"`
	VehicleType  parkingdomain.VehicleType `json:"vehicle_type"`
	LicensePlate string                    `json:"license_plate"`
	ParkingSpot  string                    `json:"parking_spot"`
}

type updateParkingRequest struct {
	VehicleType  *parkingdomain.VehicleType `json:"vehicle_type"`
	LicensePlate *string                    `json:"license_plate"`
	ParkingSpot  *string                    `json:"parking_spot"`
}

func NewHandler(usecase *parkingusecase.Service) *Handler {
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

func parseListFilters(c fiber.Ctx) (parkingusecase.ListFilters, error) {
	tenantID, err := parseUUIDQuery(c, "tenant_id")
	if err != nil {
		return parkingusecase.ListFilters{}, err
	}

	var vehicleType *parkingdomain.VehicleType
	if raw := strings.TrimSpace(c.Query("vehicle_type")); raw != "" {
		vt := parkingdomain.VehicleType(raw)
		if !vt.Valid() {
			return parkingusecase.ListFilters{}, apierror.BadRequest("invalid vehicle_type")
		}
		vehicleType = &vt
	}

	return parkingusecase.ListFilters{
		TenantID:    tenantID,
		VehicleType: vehicleType,
	}, nil
}

// List godoc
// @Summary List parking registrations
// @Description Returns tenant vehicle parking registrations. Optionally filter by tenant or vehicle type.
// @Tags parking
// @Produce json
// @Param tenant_id query string false "Filter by tenant ID"
// @Param vehicle_type query string false "Filter by vehicle type (car, motorcycle, other)"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /parking [get]
func (h *Handler) List(c fiber.Ctx) error {
	if _, ok := middleware.UserID(c); !ok {
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

	parkings, total, err := h.usecase.List(ctx, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list parking registrations")
	}

	return apiresponse.Paginated(c, parkings, page, perPage, total)
}

// Get godoc
// @Summary Get a parking registration by ID
// @Tags parking
// @Produce json
// @Param id path string true "Parking registration ID"
// @Success 200 {object} parkingdomain.Parking
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /parking/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid parking registration id")
	}

	if _, ok := middleware.UserID(c); !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	parking, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, parkingdomain.ErrParkingNotFound) {
			return apierror.NotFound("parking registration not found")
		}
		return apierror.Internal("failed to get parking registration")
	}

	return apiresponse.OK(c, parking)
}

// Create godoc
// @Summary Register a tenant vehicle for parking
// @Tags parking
// @Accept json
// @Produce json
// @Param request body createParkingRequest true "Parking registration payload"
// @Success 201 {object} parkingdomain.Parking
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /parking [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createParkingRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	parking, err := h.usecase.Create(ctx, parkingusecase.CreateInput{
		TenantID:     req.TenantID,
		VehicleType:  req.VehicleType,
		LicensePlate: req.LicensePlate,
		ParkingSpot:  req.ParkingSpot,
		CreatedBy:    &requesterID,
	})
	if err != nil {
		if errors.Is(err, parkingdomain.ErrRequiredParkingData) {
			return apierror.BadRequest("tenant_id and license_plate are required")
		}
		if errors.Is(err, parkingdomain.ErrInvalidVehicleType) {
			return apierror.BadRequest("invalid vehicle type")
		}
		if errors.Is(err, parkingdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		return apierror.Internal("failed to register parking")
	}

	return apiresponse.Created(c, parking)
}

// Update godoc
// @Summary Update a parking registration
// @Tags parking
// @Accept json
// @Produce json
// @Param id path string true "Parking registration ID"
// @Param request body updateParkingRequest true "Parking registration payload"
// @Success 200 {object} parkingdomain.Parking
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /parking/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid parking registration id")
	}

	var req updateParkingRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	parking, err := h.usecase.Update(ctx, id, parkingusecase.UpdateInput{
		VehicleType:  req.VehicleType,
		LicensePlate: req.LicensePlate,
		ParkingSpot:  req.ParkingSpot,
		UpdatedBy:    &requesterID,
	})
	if err != nil {
		if errors.Is(err, parkingdomain.ErrParkingNotFound) {
			return apierror.NotFound("parking registration not found")
		}
		if errors.Is(err, parkingdomain.ErrRequiredParkingData) {
			return apierror.BadRequest("license_plate must not be empty")
		}
		if errors.Is(err, parkingdomain.ErrInvalidVehicleType) {
			return apierror.BadRequest("invalid vehicle type")
		}
		return apierror.Internal("failed to update parking registration")
	}

	return apiresponse.OK(c, parking)
}

// Delete godoc
// @Summary Delete a parking registration
// @Tags parking
// @Produce json
// @Param id path string true "Parking registration ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /parking/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid parking registration id")
	}

	if _, ok := middleware.UserID(c); !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id); err != nil {
		if errors.Is(err, parkingdomain.ErrParkingNotFound) {
			return apierror.NotFound("parking registration not found")
		}
		return apierror.Internal("failed to delete parking registration")
	}

	return apiresponse.Message(c, "parking registration deleted")
}
