package v1

import (
	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type MenuHandler struct {
	menus *usecase.MenuUseCase
}

func NewMenuHandler(menus *usecase.MenuUseCase) *MenuHandler {
	return &MenuHandler{menus: menus}
}

func (h *MenuHandler) List(c fiber.Ctx) error {
	menus, err := h.menus.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": menus})
}

func (h *MenuHandler) GetByID(c fiber.Ctx) error {
	menu, err := h.menus.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": menu})
}
