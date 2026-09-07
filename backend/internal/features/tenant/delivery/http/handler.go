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
	LineID           string `json:"line_id"`
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
	LineID           *string `json:"line_id"`
	IDCard           *string `json:"id_card"`
	Email            *string `json:"email"`
	EmergencyContact *string `json:"emergency_contact"`
	Note             *string `json:"note"`
	IsActive         *bool   `json:"is_active"`
}

func NewHandler(usecase *tenantusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

type linkTenantLineRequest struct {
	IDToken string `json:"id_token"`
}

const defaultActiveListLimit = 50

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
		if _, ok := tenantusecase.FilterColumns[name]; !ok {
			return nil, apierror.BadRequest("unsupported filter field: " + name)
		}

		if value = strings.TrimSpace(value); value != "" {
			columns[name] = value
		}
	}

	return columns, nil
}

func parseListFilter(c fiber.Ctx) (tenantusecase.ListFilter, error) {
	isActive, err := parseOptionalBool(c, "is_active")
	if err != nil {
		return tenantusecase.ListFilter{}, err
	}

	columns, err := parseColumnFilters(c)
	if err != nil {
		return tenantusecase.ListFilter{}, err
	}

	lineLinked, err := parseOptionalBool(c, "line_linked")
	if err != nil {
		return tenantusecase.ListFilter{}, err
	}

	sortKey := tenantusecase.DefaultSortKey
	if raw := strings.TrimSpace(c.Query("sort")); raw != "" {
		if _, ok := tenantusecase.SortColumns[raw]; !ok {
			return tenantusecase.ListFilter{}, apierror.BadRequest("unsupported sort field")
		}
		sortKey = raw
	}

	// Newest-first is the useful default for an unsorted listing, but an
	// explicit sort field reads more naturally ascending.
	sortDesc := sortKey == tenantusecase.DefaultSortKey
	switch strings.TrimSpace(c.Query("order")) {
	case "":
	case "asc":
		sortDesc = false
	case "desc":
		sortDesc = true
	default:
		return tenantusecase.ListFilter{}, apierror.BadRequest("order must be asc or desc")
	}

	return tenantusecase.ListFilter{
		Search:     strings.TrimSpace(c.Query("q")),
		Columns:    columns,
		IsActive:   isActive,
		LineLinked: lineLinked,
		SortKey:    sortKey,
		SortDesc:   sortDesc,
	}, nil
}

// List godoc
// @Summary List tenants
// @Tags tenants
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Param q query string false "Filter by name, phone, LINE ID, id card or email"
// @Param f[column] query string false "Per-column substring filter, e.g. f[phone]=081; column must be one of first_name, last_name, phone, line_id, id_card, email"
// @Param is_active query bool false "Filter by active status"
// @Param line_linked query bool false "Filter by whether a LINE account is linked"
// @Param sort query string false "Sort field: first_name, last_name, phone, line_id, id_card, email, is_active, created_at (default created_at)"
// @Param order query string false "Sort direction: asc or desc"
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

	filter, err := parseListFilter(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	tenants, total, err := h.usecase.List(ctx, filter, perPage, offset)
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
		LineID:           req.LineID,
		IDCard:           req.IDCard,
		Email:            req.Email,
		EmergencyContact: req.EmergencyContact,
		Note:             req.Note,
		IsActive:         isActive,
		CreatedBy:        &requesterID,
	}, c.IP())
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
		LineID:           req.LineID,
		IDCard:           req.IDCard,
		Email:            req.Email,
		EmergencyContact: req.EmergencyContact,
		Note:             req.Note,
		IsActive:         req.IsActive,
		UpdatedBy:        &requesterID,
	}, c.IP())
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
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /tenants/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid tenant id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID, c.IP()); err != nil {
		if errors.Is(err, tenantdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		if errors.Is(err, tenantdomain.ErrTenantHasContracts) {
			return apierror.Conflict("tenant has contracts and cannot be deleted")
		}
		return apierror.Internal("failed to delete tenant")
	}

	return apiresponse.Message(c, "tenant deleted")
}

// LinkLine godoc
// @Summary Link a tenant's LINE account
// @Description Public endpoint called from the LIFF linking page: verifies the id token the tenant's LINE app produced after login and stores the resulting LINE userId on the tenant, so invoices can be pushed to them. Not authenticated, since the tenant has no login of their own — the tenant ID in the URL acts as the shared secret (only someone holding the tenant's personal linking link can call this).
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param request body linkTenantLineRequest true "LIFF id token payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Router /public/tenants/{id}/line/link [post]
func (h *Handler) LinkLine(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid tenant id")
	}

	var req linkTenantLineRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if _, err := h.usecase.LinkLine(ctx, id, req.IDToken); err != nil {
		if errors.Is(err, tenantdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		if errors.Is(err, tenantdomain.ErrInvalidLineToken) {
			return apierror.BadRequest("invalid or expired LINE id token")
		}
		if errors.Is(err, tenantdomain.ErrLineAccountAlreadyLinked) {
			return apierror.Conflict("this LINE account is already linked to another tenant").WithSlug("line_account_already_linked")
		}
		return apierror.Internal("failed to link LINE account")
	}

	return apiresponse.Message(c, "line account linked")
}

// LineOAInfo godoc
// @Summary Get the dormitory's LINE Official Account details
// @Description Public endpoint used by the LIFF linking page: returns the OA's basic ID and the URL that adds it as a friend, so a tenant who linked their account but isn't a friend yet (and so can't be pushed to) can add it in one tap. The basic ID is public information — it's on the OA's own profile and QR code.
// @Tags tenants
// @Produce json
// @Success 200 {object} tenantusecase.LineOAInfo
// @Failure 500 {object} apierror.Error
// @Router /public/line/oa [get]
func (h *Handler) LineOAInfo(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	info, err := h.usecase.LineOAInfo(ctx)
	if err != nil {
		return apierror.Internal("failed to load LINE OA info")
	}

	return apiresponse.OK(c, info)
}

// LineStatus godoc
// @Summary Check whether a tenant can be sent invoices over LINE
// @Description Reports the two independent conditions a pushed invoice needs: the tenant has linked their LINE account (so there is a userId to push to), and that account has the dormitory's OA as a friend (LINE drops messages to non-friends). Checking friendship calls LINE, so this is on-demand rather than part of the tenant list.
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} tenantusecase.LineStatus
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /tenants/{id}/line/status [get]
func (h *Handler) LineStatus(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid tenant id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	status, err := h.usecase.LineStatus(ctx, id)
	if err != nil {
		if errors.Is(err, tenantdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		return apierror.Internal("failed to check LINE status")
	}

	return apiresponse.OK(c, status)
}

// UnlinkLine godoc
// @Summary Unlink a tenant's LINE account
// @Description Clears the tenant's stored LINE userId so their personal linking link can be used again to link a (possibly different) LINE account.
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} tenantdomain.Tenant
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /tenants/{id}/line [delete]
func (h *Handler) UnlinkLine(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid tenant id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	tenant, err := h.usecase.UnlinkLine(ctx, id, requesterID, c.IP())
	if err != nil {
		if errors.Is(err, tenantdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		return apierror.Internal("failed to unlink LINE account")
	}

	return apiresponse.OK(c, tenant)
}

// CheckDeletion godoc
// @Summary Check whether a tenant can be deleted
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} tenantusecase.DeletionCheck
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /tenants/{id}/deletion-check [get]
func (h *Handler) CheckDeletion(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid tenant id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	check, err := h.usecase.CheckDeletion(ctx, id)
	if err != nil {
		if errors.Is(err, tenantdomain.ErrTenantNotFound) {
			return apierror.NotFound("tenant not found")
		}
		return apierror.Internal("failed to check tenant deletion")
	}

	return apiresponse.OK(c, check)
}
