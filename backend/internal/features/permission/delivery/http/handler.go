package http

import (
	"strings"

	permissiondomain "apihorpug/internal/features/permission/domain"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

type createPermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) List(c fiber.Ctx) error {
	var permissions []permissiondomain.Permission
	if err := h.db.Order("name asc").Find(&permissions).Error; err != nil {
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
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
	}

	if err := h.db.Create(&permission).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "permission name already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create permission"})
	}

	return c.Status(fiber.StatusCreated).JSON(permission)
}
