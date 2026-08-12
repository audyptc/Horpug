package http

import (
	"context"
	"errors"
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

// List godoc
// @Summary List roles
// @Tags roles
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /roles [get]
func (h *Handler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	roles, total, err := h.usecase.List(ctx, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list roles")
	}

	return apiresponse.Paginated(c, roles, page, perPage, total)
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
	})
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
	})
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
		return apierror.Internal("failed to update role")
	}

	return apiresponse.OK(c, updatedRole)
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

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id); err != nil {
		if errors.Is(err, roledomain.ErrRoleNotFound) {
			return apierror.NotFound("role not found")
		}
		if errors.Is(err, roledomain.ErrRoleInUse) {
			return apierror.Conflict("role is being used by users")
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
