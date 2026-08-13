package http

import (
	"context"
	"errors"
	"strings"
	"time"

	contractdomain "apihorpug/internal/features/contract/domain"
	contractusecase "apihorpug/internal/features/contract/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *contractusecase.Service
}

type createContractRequest struct {
	TenantID     uuid.UUID  `json:"tenant_id"`
	RoomID       uuid.UUID  `json:"room_id"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	RentPrice    float64    `json:"rent_price"`
	Deposit      float64    `json:"deposit"`
	NumOccupants int        `json:"num_occupants"`
	Note         string     `json:"note"`
}

type updateContractRequest struct {
	EndDate      *time.Time                     `json:"end_date"`
	RentPrice    *float64                       `json:"rent_price"`
	Deposit      *float64                       `json:"deposit"`
	NumOccupants *int                           `json:"num_occupants"`
	Status       *contractdomain.ContractStatus `json:"status"`
	Note         *string                        `json:"note"`
}

func NewHandler(usecase *contractusecase.Service) *Handler {
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

func parseListFilters(c fiber.Ctx) (contractusecase.ListFilters, error) {
	tenantID, err := parseUUIDQuery(c, "tenant_id")
	if err != nil {
		return contractusecase.ListFilters{}, err
	}
	roomID, err := parseUUIDQuery(c, "room_id")
	if err != nil {
		return contractusecase.ListFilters{}, err
	}
	dormitoryID, err := parseUUIDQuery(c, "dormitory_id")
	if err != nil {
		return contractusecase.ListFilters{}, err
	}

	var status *contractdomain.ContractStatus
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		s := contractdomain.ContractStatus(raw)
		if !s.Valid() {
			return contractusecase.ListFilters{}, apierror.BadRequest("invalid status")
		}
		status = &s
	}

	return contractusecase.ListFilters{
		TenantID:    tenantID,
		RoomID:      roomID,
		DormitoryID: dormitoryID,
		Status:      status,
	}, nil
}

// List godoc
// @Summary List contracts
// @Description Returns every contract for roles with full dormitory access, otherwise only contracts under dormitories the caller manages. Optionally filter by tenant, room, dormitory or status.
// @Tags contracts
// @Produce json
// @Param tenant_id query string false "Filter by tenant ID"
// @Param room_id query string false "Filter by room ID"
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param status query string false "Filter by status (active, expired, terminated)"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /contracts [get]
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

	contracts, total, err := h.usecase.List(ctx, requesterID, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list contracts")
	}

	return apiresponse.Paginated(c, contracts, page, perPage, total)
}

// Get godoc
// @Summary Get a contract by ID
// @Tags contracts
// @Produce json
// @Param id path string true "Contract ID"
// @Success 200 {object} contractdomain.Contract
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /contracts/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid contract id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	contract, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, contractdomain.ErrContractNotFound) {
			return apierror.NotFound("contract not found")
		}
		return apierror.Internal("failed to get contract")
	}

	return apiresponse.OK(c, contract)
}

// Create godoc
// @Summary Create a contract
// @Description Creates a lease contract linking a tenant to a room. A room may only have one active contract at a time.
// @Tags contracts
// @Accept json
// @Produce json
// @Param request body createContractRequest true "Contract payload"
// @Success 201 {object} contractdomain.Contract
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /contracts [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createContractRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	contract, err := h.usecase.Create(ctx, contractusecase.CreateInput{
		TenantID:     req.TenantID,
		RoomID:       req.RoomID,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		RentPrice:    req.RentPrice,
		Deposit:      req.Deposit,
		NumOccupants: req.NumOccupants,
		Note:         req.Note,
		CreatedBy:    &requesterID,
	})
	if err != nil {
		if errors.Is(err, contractdomain.ErrRequiredContractData) {
			return apierror.BadRequest("tenant_id, room_id and start_date are required")
		}
		if errors.Is(err, contractdomain.ErrInvalidContractDates) {
			return apierror.BadRequest("end_date must not be before start_date")
		}
		if errors.Is(err, contractdomain.ErrInvalidContractAmount) {
			return apierror.BadRequest("rent_price and deposit must not be negative")
		}
		if errors.Is(err, contractdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		if errors.Is(err, contractdomain.ErrRoomNotFound) {
			return apierror.NotFound("room not found")
		}
		if errors.Is(err, contractdomain.ErrRoomHasActiveContract) {
			return apierror.Conflict("room already has an active contract")
		}
		return apierror.Internal("failed to create contract")
	}

	return apiresponse.Created(c, contract)
}

// Update godoc
// @Summary Update a contract
// @Tags contracts
// @Accept json
// @Produce json
// @Param id path string true "Contract ID"
// @Param request body updateContractRequest true "Contract payload"
// @Success 200 {object} contractdomain.Contract
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /contracts/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid contract id")
	}

	var req updateContractRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	contract, err := h.usecase.Update(ctx, id, requesterID, contractusecase.UpdateInput{
		EndDate:      req.EndDate,
		RentPrice:    req.RentPrice,
		Deposit:      req.Deposit,
		NumOccupants: req.NumOccupants,
		Status:       req.Status,
		Note:         req.Note,
		UpdatedBy:    &requesterID,
	})
	if err != nil {
		if errors.Is(err, contractdomain.ErrContractNotFound) {
			return apierror.NotFound("contract not found")
		}
		if errors.Is(err, contractdomain.ErrInvalidContractAmount) {
			return apierror.BadRequest("rent_price and deposit must not be negative")
		}
		if errors.Is(err, contractdomain.ErrInvalidNumOccupants) {
			return apierror.BadRequest("num_occupants must be at least 1")
		}
		if errors.Is(err, contractdomain.ErrInvalidContractStatus) {
			return apierror.BadRequest("invalid status")
		}
		if errors.Is(err, contractdomain.ErrRoomHasActiveContract) {
			return apierror.Conflict("room already has an active contract")
		}
		return apierror.Internal("failed to update contract")
	}

	return apiresponse.OK(c, contract)
}

// Delete godoc
// @Summary Delete a contract
// @Tags contracts
// @Produce json
// @Param id path string true "Contract ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /contracts/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid contract id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID); err != nil {
		if errors.Is(err, contractdomain.ErrContractNotFound) {
			return apierror.NotFound("contract not found")
		}
		return apierror.Internal("failed to delete contract")
	}

	return apiresponse.Message(c, "contract deleted")
}
