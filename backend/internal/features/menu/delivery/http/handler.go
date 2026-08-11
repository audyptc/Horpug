package http

import (
	"context"
	"time"

	menudomain "apihorpug/internal/features/menu/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.db.Query(ctx, `
		SELECT id, name, path, description, is_active, created_at, updated_at
		FROM menus
		ORDER BY path ASC
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list menus"})
	}
	defer rows.Close()

	menus := make([]menudomain.Menu, 0)
	for rows.Next() {
		var menu menudomain.Menu
		if err := rows.Scan(&menu.ID, &menu.Name, &menu.Path, &menu.Description, &menu.IsActive, &menu.CreatedAt, &menu.UpdatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list menus"})
		}
		menus = append(menus, menu)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list menus"})
	}

	return c.JSON(menus)
}
