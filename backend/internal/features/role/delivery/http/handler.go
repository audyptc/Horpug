package http

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	roledomain "apihorpug/internal/features/role/domain"
	roleusecase "apihorpug/internal/features/role/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *roleusecase.Service
}

type menuPermissionInput struct {
	MenuID        uuid.UUID   `json:"menu_id"`
	PermissionIDs []uuid.UUID `json:"permission_ids"`
}

type createRoleRequest struct {
	Name                string                `json:"name"`
	Description         string                `json:"description"`
	IsActive            *bool                 `json:"is_active"`
	FullDormitoryAccess *bool                 `json:"full_dormitory_access"`
	DormitoryIDs        []uuid.UUID           `json:"dormitory_ids"`
	MenuPermissions     []menuPermissionInput `json:"menu_permissions"`
}

type updateRoleRequest struct {
	Name                *string                `json:"name"`
	Description         *string                `json:"description"`
	IsActive            *bool                  `json:"is_active"`
	FullDormitoryAccess *bool                  `json:"full_dormitory_access"`
	DormitoryIDs        *[]uuid.UUID           `json:"dormitory_ids"`
	MenuPermissions     *[]menuPermissionInput `json:"menu_permissions"`
}

func NewHandler(usecase *roleusecase.Service) *Handler {
	return &Handler{usecase: usecase}
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
		if _, ok := roleusecase.FilterColumns[name]; !ok {
			return nil, apierror.BadRequest("unsupported filter field: " + name)
		}

		if value = strings.TrimSpace(value); value != "" {
			columns[name] = value
		}
	}

	return columns, nil
}

func parseListFilters(c fiber.Ctx) (roleusecase.ListFilters, error) {
	isActive, err := parseOptionalBool(c, "is_active")
	if err != nil {
		return roleusecase.ListFilters{}, err
	}

	columns, err := parseColumnFilters(c)
	if err != nil {
		return roleusecase.ListFilters{}, err
	}

	sortKey := roleusecase.DefaultSortKey
	if raw := strings.TrimSpace(c.Query("sort")); raw != "" {
		if _, ok := roleusecase.SortColumns[raw]; !ok {
			return roleusecase.ListFilters{}, apierror.BadRequest("unsupported sort field")
		}
		sortKey = raw
	}

	// Role names read most naturally sorted ascending, so unlike lists
	// defaulting to a newest-first field, no key flips the default direction.
	sortDesc := false
	switch strings.TrimSpace(c.Query("order")) {
	case "":
	case "asc":
		sortDesc = false
	case "desc":
		sortDesc = true
	default:
		return roleusecase.ListFilters{}, apierror.BadRequest("order must be asc or desc")
	}

	return roleusecase.ListFilters{
		Search:   strings.TrimSpace(c.Query("q")),
		Columns:  columns,
		IsActive: isActive,
		SortKey:  sortKey,
		SortDesc: sortDesc,
	}, nil
}

// List godoc
// @Summary List roles
// @Tags roles
// @Produce json
// @Param q query string false "Filter by name or description"
// @Param f[column] query string false "Per-column substring filter, e.g. f[name]=manager; column must be one of name, description"
// @Param is_active query bool false "Filter by active status"
// @Param sort query string false "Sort field: name, description, is_active, created_at (default name)"
// @Param order query string false "Sort direction: asc or desc"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /roles [get]
func (h *Handler) List(c fiber.Ctx) error {
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

	roles, total, err := h.usecase.List(ctx, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list roles")
	}

	return apiresponse.Paginated(c, roles, page, perPage, total)
}

// ListActive godoc
// @Summary List active roles
// @Description Returns active roles. Intended for populating role selectors; results are capped rather than paginated.
// @Tags roles
// @Produce json
// @Param q query string false "Filter by name"
// @Param limit query int false "Max results (default 50, max 100)"
// @Success 200 {array} roledomain.Role
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /roles/active [get]
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

	roles, err := h.usecase.ListActive(ctx, search, limit)
	if err != nil {
		return apierror.Internal("failed to list active roles")
	}

	return apiresponse.OK(c, roles)
}

// Get godoc
// @Summary Get a role by ID
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} roledomain.Role
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /roles/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid role id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	role, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, roledomain.ErrRoleNotFound) {
			return apierror.NotFound("role not found")
		}
		return apierror.Internal("failed to get role")
	}

	return apiresponse.OK(c, role)
}

// Create godoc
// @Summary Create a role
// @Tags roles
// @Accept json
// @Produce json
// @Param request body createRoleRequest true "Role payload"
// @Success 201 {object} roledomain.Role
// @Failure 400 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /roles [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return apierror.BadRequest("name is required")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	fullDormitoryAccess := false
	if req.FullDormitoryAccess != nil {
		fullDormitoryAccess = *req.FullDormitoryAccess
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	createdRole, err := h.usecase.Create(ctx, roleusecase.CreateInput{
		Name:                req.Name,
		Description:         req.Description,
		IsActive:            isActive,
		FullDormitoryAccess: fullDormitoryAccess,
		DormitoryIDs:        req.DormitoryIDs,
		MenuPermissions:     toUsecaseMenuPermissions(req.MenuPermissions),
		CreatedBy:           &requesterID,
	}, c.IP())
	if err != nil {
		if errors.Is(err, roledomain.ErrRoleNameExists) {
			return apierror.Conflict("role name already exists")
		}
		if errors.Is(err, roledomain.ErrReferenceNotFound) {
			return apierror.BadRequest("one or more menus, permissions or dormitories not found")
		}
		return apierror.Internal("failed to create role")
	}

	return apiresponse.Created(c, createdRole)
}

// Update godoc
// @Summary Update a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body updateRoleRequest true "Role payload"
// @Success 200 {object} roledomain.Role
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /roles/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid role id")
	}

	var req updateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return apierror.BadRequest("name cannot be empty")
		}
		req.Name = &name
	}

	var menuPermissions *[]roleusecase.MenuPermissionInput
	if req.MenuPermissions != nil {
		mapped := toUsecaseMenuPermissions(*req.MenuPermissions)
		menuPermissions = &mapped
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	updatedRole, err := h.usecase.Update(ctx, id, roleusecase.UpdateInput{
		Name:                req.Name,
		Description:         req.Description,
		IsActive:            req.IsActive,
		FullDormitoryAccess: req.FullDormitoryAccess,
		DormitoryIDs:        req.DormitoryIDs,
		MenuPermissions:     menuPermissions,
		UpdatedBy:           &requesterID,
	}, c.IP())
	if err != nil {
		if errors.Is(err, roledomain.ErrRoleNotFound) {
			return apierror.NotFound("role not found")
		}
		if errors.Is(err, roledomain.ErrRoleNameExists) {
			return apierror.Conflict("role name already exists")
		}
		if errors.Is(err, roledomain.ErrReferenceNotFound) {
			return apierror.BadRequest("one or more menus, permissions or dormitories not found")
		}
		if errors.Is(err, roledomain.ErrRoleProtected) {
			return apierror.Forbidden("this role is protected and cannot be modified")
		}
		return apierror.Internal("failed to update role")
	}

	return apiresponse.OK(c, updatedRole)
}

// CheckDeletion godoc
// @Summary Check whether a role can be deleted
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} roleusecase.DeletionCheck
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /roles/{id}/deletion-check [get]
func (h *Handler) CheckDeletion(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid role id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	check, err := h.usecase.CheckDeletion(ctx, id)
	if err != nil {
		if errors.Is(err, roledomain.ErrRoleNotFound) {
			return apierror.NotFound("role not found")
		}
		return apierror.Internal("failed to check role deletion")
	}

	return apiresponse.OK(c, check)
}

// Delete godoc
// @Summary Delete a role
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /roles/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid role id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID, c.IP()); err != nil {
		if errors.Is(err, roledomain.ErrRoleNotFound) {
			return apierror.NotFound("role not found")
		}
		if errors.Is(err, roledomain.ErrRoleInUse) {
			return apierror.Conflict("role is being used by users")
		}
		if errors.Is(err, roledomain.ErrRoleProtected) {
			return apierror.Forbidden("this role is protected and cannot be deleted")
		}
		return apierror.Internal("failed to delete role")
	}

	return apiresponse.Message(c, "role deleted")
}

func toUsecaseMenuPermissions(inputs []menuPermissionInput) []roleusecase.MenuPermissionInput {
	results := make([]roleusecase.MenuPermissionInput, 0, len(inputs))
	for _, input := range inputs {
		results = append(results, roleusecase.MenuPermissionInput{
			MenuID:        input.MenuID,
			PermissionIDs: input.PermissionIDs,
		})
	}
	return results
}
