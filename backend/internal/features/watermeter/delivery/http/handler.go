package http

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	metdomain "apihorpug/internal/features/watermeter/domain"
	metusecase "apihorpug/internal/features/watermeter/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *metusecase.Service
}

type createMeterRequest struct {
	RoomID        uuid.UUID               `json:"room_id"`
	BillingMethod metdomain.BillingMethod `json:"billing_method"`
	ReadingDate   time.Time               `json:"reading_date"`
	PreviousUnit  float64                 `json:"previous_unit"`
	CurrentUnit   float64                 `json:"current_unit"`
	PricePerUnit  float64                 `json:"price_per_unit"`
	FlatAmount    *float64                `json:"flat_amount"`
	Note          string                  `json:"note"`
}

type updateMeterRequest struct {
	BillingMethod *metdomain.BillingMethod `json:"billing_method"`
	ReadingDate   *time.Time               `json:"reading_date"`
	PreviousUnit  *float64                 `json:"previous_unit"`
	CurrentUnit   *float64                 `json:"current_unit"`
	PricePerUnit  *float64                 `json:"price_per_unit"`
	FlatAmount    *float64                 `json:"flat_amount"`
	Note          *string                  `json:"note"`
}

func NewHandler(usecase *metusecase.Service) *Handler {
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

// parseOptionalBool reads a tri-state query flag: absent means "don't filter
// on this", so an unset value is distinct from an explicit false.
func parseOptionalBool(c fiber.Ctx, name string) (*bool, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, apierror.BadRequest(name + " must be true or false")
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
		if _, ok := metusecase.FilterColumns[name]; !ok {
			return nil, apierror.BadRequest("unsupported filter field: " + name)
		}

		if value = strings.TrimSpace(value); value != "" {
			columns[name] = value
		}
	}

	return columns, nil
}

func parseListFilters(c fiber.Ctx) (metusecase.ListFilters, error) {
	roomID, err := parseUUIDQuery(c, "room_id")
	if err != nil {
		return metusecase.ListFilters{}, err
	}
	dormitoryID, err := parseUUIDQuery(c, "dormitory_id")
	if err != nil {
		return metusecase.ListFilters{}, err
	}

	var billingMethod *metdomain.BillingMethod
	if raw := strings.TrimSpace(c.Query("billing_method")); raw != "" {
		method := metdomain.BillingMethod(raw)
		if !method.Valid() {
			return metusecase.ListFilters{}, apierror.BadRequest("invalid billing_method")
		}
		billingMethod = &method
	}

	isBilled, err := parseOptionalBool(c, "is_billed")
	if err != nil {
		return metusecase.ListFilters{}, err
	}

	columns, err := parseColumnFilters(c)
	if err != nil {
		return metusecase.ListFilters{}, err
	}

	sortKey := metusecase.DefaultSortKey
	if raw := strings.TrimSpace(c.Query("sort")); raw != "" {
		if _, ok := metusecase.SortColumns[raw]; !ok {
			return metusecase.ListFilters{}, apierror.BadRequest("unsupported sort field")
		}
		sortKey = raw
	}

	// Newest reading first is the useful default for an unsorted listing, but
	// an explicit sort field reads more naturally ascending.
	sortDesc := sortKey == metusecase.DefaultSortKey
	switch strings.TrimSpace(c.Query("order")) {
	case "":
	case "asc":
		sortDesc = false
	case "desc":
		sortDesc = true
	default:
		return metusecase.ListFilters{}, apierror.BadRequest("order must be asc or desc")
	}

	return metusecase.ListFilters{
		RoomID:        roomID,
		DormitoryID:   dormitoryID,
		BillingMethod: billingMethod,
		IsBilled:      isBilled,
		Search:        strings.TrimSpace(c.Query("q")),
		Columns:       columns,
		SortKey:       sortKey,
		SortDesc:      sortDesc,
	}, nil
}

// List godoc
// @Summary List water meter readings
// @Description Returns every water meter reading for roles with full dormitory access, otherwise only readings under dormitories the caller manages. Optionally filter by room or dormitory.
// @Tags water-meters
// @Produce json
// @Param room_id query string false "Filter by room ID"
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param billing_method query string false "Filter by billing method (metered, flat)"
// @Param is_billed query bool false "Filter by whether the reading has been billed on an invoice"
// @Param q query string false "Filter by room number or dormitory name"
// @Param f[column] query string false "Per-column substring filter, e.g. f[room_number]=101; column must be one of room_number, dormitory_name"
// @Param sort query string false "Sort field: room_number, dormitory_name, reading_date, billing_method, previous_unit, current_unit, unit_used, price_per_unit, total_amount, is_billed, created_at (default reading_date)"
// @Param order query string false "Sort direction: asc or desc"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /water-meters [get]
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

	meters, total, err := h.usecase.List(ctx, requesterID, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list water meter readings")
	}

	return apiresponse.Paginated(c, meters, page, perPage, total)
}

// Get godoc
// @Summary Get a water meter reading by ID
// @Tags water-meters
// @Produce json
// @Param id path string true "Meter reading ID"
// @Success 200 {object} metdomain.Meter
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /water-meters/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid meter reading id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	meter, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, metdomain.ErrMeterNotFound) {
			return apierror.NotFound("water meter reading not found")
		}
		return apierror.Internal("failed to get water meter reading")
	}

	return apiresponse.OK(c, meter)
}

// Create godoc
// @Summary Record a water meter reading
// @Description Records a water meter reading or charge for a room. billing_method "metered" (default) derives total_amount as (current_unit - previous_unit) * price_per_unit; billing_method "flat" uses flat_amount directly as total_amount, for flat-rate, tiered or shared-meter charges computed externally.
// @Tags water-meters
// @Accept json
// @Produce json
// @Param request body createMeterRequest true "Meter reading payload"
// @Success 201 {object} metdomain.Meter
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /water-meters [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createMeterRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	meter, err := h.usecase.Create(ctx, metusecase.CreateInput{
		RoomID:        req.RoomID,
		BillingMethod: req.BillingMethod,
		ReadingDate:   req.ReadingDate,
		PreviousUnit:  req.PreviousUnit,
		CurrentUnit:   req.CurrentUnit,
		PricePerUnit:  req.PricePerUnit,
		FlatAmount:    req.FlatAmount,
		Note:          req.Note,
		CreatedBy:     &requesterID,
	}, c.IP())
	if err != nil {
		if errors.Is(err, metdomain.ErrRequiredMeterData) {
			return apierror.BadRequest("room_id and reading_date are required")
		}
		if errors.Is(err, metdomain.ErrInvalidBillingMethod) {
			return apierror.BadRequest("invalid billing_method")
		}
		if errors.Is(err, metdomain.ErrInvalidMeterUnits) {
			return apierror.BadRequest("previous_unit and current_unit must not be negative, and current_unit must not be less than previous_unit")
		}
		if errors.Is(err, metdomain.ErrInvalidMeterPrice) {
			return apierror.BadRequest("price_per_unit must not be negative")
		}
		if errors.Is(err, metdomain.ErrRequiredFlatAmount) {
			return apierror.BadRequest("flat_amount is required and must not be negative when billing_method is flat")
		}
		if errors.Is(err, metdomain.ErrRoomNotFound) {
			return apierror.NotFound("room not found")
		}
		if errors.Is(err, metdomain.ErrMeterReadingExists) {
			return apierror.Conflict("a water meter reading for this room and date already exists")
		}
		if errors.Is(err, metdomain.ErrMeterMonthExists) {
			return apierror.Conflict("a water meter reading for this room and month already exists")
		}
		return apierror.Internal("failed to record water meter reading")
	}

	return apiresponse.Created(c, meter)
}

// Update godoc
// @Summary Update a water meter reading
// @Tags water-meters
// @Accept json
// @Produce json
// @Param id path string true "Meter reading ID"
// @Param request body updateMeterRequest true "Meter reading payload"
// @Success 200 {object} metdomain.Meter
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /water-meters/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid meter reading id")
	}

	var req updateMeterRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	meter, err := h.usecase.Update(ctx, id, requesterID, metusecase.UpdateInput{
		BillingMethod: req.BillingMethod,
		ReadingDate:   req.ReadingDate,
		PreviousUnit:  req.PreviousUnit,
		CurrentUnit:   req.CurrentUnit,
		PricePerUnit:  req.PricePerUnit,
		FlatAmount:    req.FlatAmount,
		Note:          req.Note,
		UpdatedBy:     &requesterID,
	}, c.IP())
	if err != nil {
		if errors.Is(err, metdomain.ErrMeterNotFound) {
			return apierror.NotFound("water meter reading not found")
		}
		if errors.Is(err, metdomain.ErrInvalidBillingMethod) {
			return apierror.BadRequest("invalid billing_method")
		}
		if errors.Is(err, metdomain.ErrInvalidMeterUnits) {
			return apierror.BadRequest("previous_unit and current_unit must not be negative, and current_unit must not be less than previous_unit")
		}
		if errors.Is(err, metdomain.ErrInvalidMeterPrice) {
			return apierror.BadRequest("price_per_unit must not be negative")
		}
		if errors.Is(err, metdomain.ErrRequiredFlatAmount) {
			return apierror.BadRequest("flat_amount is required and must not be negative when billing_method is flat")
		}
		if errors.Is(err, metdomain.ErrMeterReadingExists) {
			return apierror.Conflict("a water meter reading for this room and date already exists")
		}
		if errors.Is(err, metdomain.ErrMeterMonthExists) {
			return apierror.Conflict("a water meter reading for this room and month already exists")
		}
		return apierror.Internal("failed to update water meter reading")
	}

	return apiresponse.OK(c, meter)
}

// Delete godoc
// @Summary Delete a water meter reading
// @Tags water-meters
// @Produce json
// @Param id path string true "Meter reading ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /water-meters/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid meter reading id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID, c.IP()); err != nil {
		if errors.Is(err, metdomain.ErrMeterNotFound) {
			return apierror.NotFound("water meter reading not found")
		}
		if errors.Is(err, metdomain.ErrMeterHasBilledUsage) {
			return apierror.Conflict("water meter reading has already been billed and cannot be deleted")
		}
		return apierror.Internal("failed to delete water meter reading")
	}

	return apiresponse.Message(c, "water meter reading deleted")
}
