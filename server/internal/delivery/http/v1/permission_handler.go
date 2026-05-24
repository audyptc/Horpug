package v1

import (
	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type PermissionHandler struct {
	perms *usecase.PermissionUseCase
}

func NewPermissionHandler(perms *usecase.PermissionUseCase) *PermissionHandler {
	return &PermissionHandler{perms: perms}
}

func (h *PermissionHandler) List(c fiber.Ctx) error {
	perms, err := h.perms.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": perms})
}
