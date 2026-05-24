package v1

import (
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type RoleHandler struct {
	roles *usecase.RoleUseCase
}

func NewRoleHandler(roles *usecase.RoleUseCase) *RoleHandler {
	return &RoleHandler{roles: roles}
}

func (h *RoleHandler) List(c fiber.Ctx) error {
	roles, err := h.roles.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": roles})
}

func (h *RoleHandler) GetByID(c fiber.Ctx) error {
	role, err := h.roles.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": role})
}

func (h *RoleHandler) Create(c fiber.Ctx) error {
	var req domain.CreateRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": "invalid request body",
		})
	}
	role, err := h.roles.Create(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": role})
}

func (h *RoleHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": "invalid request body",
		})
	}
	role, err := h.roles.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": role})
}

func (h *RoleHandler) Delete(c fiber.Ctx) error {
	if err := h.roles.Delete(c.Context(), c.Params("id")); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "message": "role deleted"})
}

func (h *RoleHandler) AssignPermissions(c fiber.Ctx) error {
	var req domain.AssignPermissionsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": "invalid request body",
		})
	}
	if err := h.roles.AssignPermissions(c.Context(), c.Params("id"), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "message": "permissions assigned successfully"})
}
