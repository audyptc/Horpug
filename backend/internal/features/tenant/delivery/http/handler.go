package http

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	tenantdomain "apihorpug/internal/features/tenant/domain"
	tenantusecase "apihorpug/internal/features/tenant/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *tenantusecase.Service
}

type createTenantRequest struct {
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Phone            string `json:"phone"`
	IDCard           string `json:"id_card"`
	Email            string `json:"email"`
	EmergencyContact string `json:"emergency_contact"`
	Note             string `json:"note"`
	IsActive         *bool  `json:"is_active"`
}

type updateTenantRequest struct {
	FirstName        *string `json:"first_name"`
	LastName         *string `json:"last_name"`
	Phone            *string `json:"phone"`
	IDCard           *string `json:"id_card"`
	Email            *string `json:"email"`
	EmergencyContact *string `json:"emergency_contact"`
	Note             *string `json:"note"`
	IsActive         *bool   `json:"is_active"`
}

func NewHandler(usecase *tenantusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

const defaultActiveListLimit = 50

// List godoc
// @Summary List tenants
// @Tags tenants
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /tenants [get]
func (h *Handler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	tenants, total, err := h.usecase.List(ctx, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list tenants")
	}

	return apiresponse.Paginated(c, tenants, page, perPage, total)
}

// ListActive godoc
// @Summary List active tenants
// @Description Returns active tenants. Intended for populating tenant selectors; results are capped rather than paginated.
// @Tags tenants
// @Produce json
// @Param q query string false "Filter by name, phone or id card"
// @Param limit query int false "Max results (default 50, max 100)"
// @Success 200 {array} tenantdomain.Tenant
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /tenants/active [get]
func (h *Handler) ListActive(c fiber.Ctx) error {
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

	tenants, err := h.usecase.ListActive(ctx, search, limit)
	if err != nil {
		return apierror.Internal("failed to list active tenants")
	}

	return apiresponse.OK(c, tenants)
}

// Get godoc
// @Summary Get a tenant by ID
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} tenantdomain.Tenant
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /tenants/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid tenant id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	tenant, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, tenantdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		return apierror.Internal("failed to get tenant")
	}

	return apiresponse.OK(c, tenant)
}

// Create godoc
// @Summary Create a tenant
// @Tags tenants
// @Accept json
// @Produce json
// @Param request body createTenantRequest true "Tenant payload"
// @Success 201 {object} tenantdomain.Tenant
// @Failure 400 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /tenants [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createTenantRequest
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

	tenant, err := h.usecase.Create(ctx, tenantusecase.CreateInput{
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Phone:            req.Phone,
		IDCard:           req.IDCard,
		Email:            req.Email,
		EmergencyContact: req.EmergencyContact,
		Note:             req.Note,
		IsActive:         isActive,
		CreatedBy:        &requesterID,
	})
	if err != nil {
		if errors.Is(err, tenantdomain.ErrRequiredTenantData) {
			return apierror.BadRequest("first_name and last_name are required")
		}
		if errors.Is(err, tenantdomain.ErrTenantIDCardExists) {
			return apierror.Conflict("id card already exists")
		}
		return apierror.Internal("failed to create tenant")
	}

	return apiresponse.Created(c, tenant)
}

// Update godoc
// @Summary Update a tenant
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param request body updateTenantRequest true "Tenant payload"
// @Success 200 {object} tenantdomain.Tenant
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /tenants/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid tenant id")
	}

	var req updateTenantRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	tenant, err := h.usecase.Update(ctx, id, tenantusecase.UpdateInput{
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Phone:            req.Phone,
		IDCard:           req.IDCard,
		Email:            req.Email,
		EmergencyContact: req.EmergencyContact,
		Note:             req.Note,
		IsActive:         req.IsActive,
		UpdatedBy:        &requesterID,
	})
	if err != nil {
		if errors.Is(err, tenantdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		if errors.Is(err, tenantdomain.ErrRequiredTenantData) {
			return apierror.BadRequest("first_name and last_name cannot be empty")
		}
		if errors.Is(err, tenantdomain.ErrTenantIDCardExists) {
			return apierror.Conflict("id card already exists")
		}
		return apierror.Internal("failed to update tenant")
	}

	return apiresponse.OK(c, tenant)
}

// Delete godoc
// @Summary Delete a tenant
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /tenants/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid tenant id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id); err != nil {
		if errors.Is(err, tenantdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		return apierror.Internal("failed to delete tenant")
	}

	return apiresponse.Message(c, "tenant deleted")
}
