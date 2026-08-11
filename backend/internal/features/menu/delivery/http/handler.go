package http

import (
	menudomain "apihorpug/internal/features/menu/domain"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) List(c fiber.Ctx) error {
	var menus []menudomain.Menu
	if err := h.db.Order("path asc").Find(&menus).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list menus"})
	}
	return c.JSON(menus)
}
