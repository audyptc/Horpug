package http

import (
	"context"
	"errors"
	"strings"
	"time"

	roledomain "apihorpug/internal/features/role/domain"
	roleusecase "apihorpug/internal/features/role/usecase"

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
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	IsActive        *bool                 `json:"is_active"`
	MenuPermissions []menuPermissionInput `json:"menu_permissions"`
}

type updateRoleRequest struct {
	Name            *string                `json:"name"`
	Description     *string                `json:"description"`
	IsActive        *bool                  `json:"is_active"`
	MenuPermissions *[]menuPermissionInput `json:"menu_permissions"`
}

func NewHandler(usecase *roleusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	roles, err := h.usecase.List(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list roles"})
	}

	return c.JSON(roles)
}

func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	role, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, roledomain.ErrRoleNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "role not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get role"})
	}

	return c.JSON(role)
}

func (h *Handler) Create(c fiber.Ctx) error {
	var req createRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	createdRole, err := h.usecase.Create(ctx, roleusecase.CreateInput{
		Name:            req.Name,
		Description:     req.Description,
		IsActive:        isActive,
		MenuPermissions: toUsecaseMenuPermissions(req.MenuPermissions),
	})
	if err != nil {
		if errors.Is(err, roledomain.ErrRoleNameExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "role name already exists"})
		}
		if errors.Is(err, roledomain.ErrReferenceNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "one or more menus or permissions not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create role"})
	}

	return c.Status(fiber.StatusCreated).JSON(createdRole)
}

func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role id"})
	}

	var req updateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name cannot be empty"})
		}
		req.Name = &name
	}

	var menuPermissions *[]roleusecase.MenuPermissionInput
	if req.MenuPermissions != nil {
		mapped := toUsecaseMenuPermissions(*req.MenuPermissions)
		menuPermissions = &mapped
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	updatedRole, err := h.usecase.Update(ctx, id, roleusecase.UpdateInput{
		Name:            req.Name,
		Description:     req.Description,
		IsActive:        req.IsActive,
		MenuPermissions: menuPermissions,
	})
	if err != nil {
		if errors.Is(err, roledomain.ErrRoleNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "role not found"})
		}
		if errors.Is(err, roledomain.ErrRoleNameExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "role name already exists"})
		}
		if errors.Is(err, roledomain.ErrReferenceNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "one or more menus or permissions not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update role"})
	}

	return c.JSON(updatedRole)
}

func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id); err != nil {
		if errors.Is(err, roledomain.ErrRoleNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "role not found"})
		}
		if errors.Is(err, roledomain.ErrRoleInUse) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "role is being used by users"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete role"})
	}

	return c.JSON(fiber.Map{"message": "role deleted"})
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
