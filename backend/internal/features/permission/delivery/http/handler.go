package http

import (
	"context"
	"errors"
	"strings"
	"time"

	permissiondomain "apihorpug/internal/features/permission/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

type createPermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.db.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM permissions
		ORDER BY name ASC
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list permissions"})
	}
	defer rows.Close()

	permissions := make([]permissiondomain.Permission, 0)
	for rows.Next() {
		var permission permissiondomain.Permission
		if err := rows.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt, &permission.UpdatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list permissions"})
		}
		permissions = append(permissions, permission)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list permissions"})
	}

	return c.JSON(permissions)
}

func (h *Handler) Create(c fiber.Ctx) error {
	var req createPermissionRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	permission := permissiondomain.Permission{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	err := h.db.QueryRow(ctx, `
		INSERT INTO permissions (id, name, description)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at
	`, permission.ID, permission.Name, permission.Description).Scan(&permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "permission name already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create permission"})
	}

	return c.Status(fiber.StatusCreated).JSON(permission)
}
