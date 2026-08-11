package http

import (
	"context"
	"time"

	menuusecase "apihorpug/internal/features/menu/usecase"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	usecase *menuusecase.Service
}

func NewHandler(usecase *menuusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	menus, err := h.usecase.List(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list menus"})
	}

	return c.JSON(menus)
}
