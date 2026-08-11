package http

import (
	"errors"
	"strings"

	menudomain "apihorpug/internal/features/menu/domain"
	permissiondomain "apihorpug/internal/features/permission/domain"
	roledomain "apihorpug/internal/features/role/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
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

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) List(c fiber.Ctx) error {
	var roles []roledomain.Role
	if err := h.db.
		Preload("MenuPermissions.Menu").
		Preload("MenuPermissions.Permission").
		Order("name asc").
		Find(&roles).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list roles"})
	}
	return c.JSON(roles)
}

func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role id"})
	}

	var role roledomain.Role
	if err := h.db.
		Preload("MenuPermissions.Menu").
		Preload("MenuPermissions.Permission").
		First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	role := roledomain.Role{
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
		IsActive:    true,
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		return replaceRoleMenuPermissions(tx, role.ID, req.MenuPermissions)
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "role name already exists"})
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "one or more menus or permissions not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create role"})
	}

	if err := h.db.
		Preload("MenuPermissions.Menu").
		Preload("MenuPermissions.Permission").
		First(&role, "id = ?", role.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load role"})
	}

	return c.Status(fiber.StatusCreated).JSON(role)
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

	var role roledomain.Role
	if err := h.db.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "role not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get role"})
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{}
		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				return fiber.NewError(fiber.StatusBadRequest, "name cannot be empty")
			}
			updates["name"] = name
		}
		if req.Description != nil {
			updates["description"] = strings.TrimSpace(*req.Description)
		}
		if req.IsActive != nil {
			updates["is_active"] = *req.IsActive
		}
		if len(updates) > 0 {
			if err := tx.Model(&role).Updates(updates).Error; err != nil {
				return err
			}
		}

		if req.MenuPermissions != nil {
			if err := replaceRoleMenuPermissions(tx, role.ID, *req.MenuPermissions); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return c.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
		}
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "role name already exists"})
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "one or more menus or permissions not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update role"})
	}

	if err := h.db.
		Preload("MenuPermissions.Menu").
		Preload("MenuPermissions.Permission").
		First(&role, "id = ?", role.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load role"})
	}

	return c.JSON(role)
}

func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role id"})
	}

	result := h.db.Delete(&roledomain.Role{}, "id = ?", id)
	if result.Error != nil {
		if strings.Contains(strings.ToLower(result.Error.Error()), "foreign key") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "role is being used by users"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete role"})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "role not found"})
	}

	return c.JSON(fiber.Map{"message": "role deleted"})
}

func replaceRoleMenuPermissions(tx *gorm.DB, roleID uuid.UUID, menuPermissions []menuPermissionInput) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&roledomain.RoleMenuPermission{}).Error; err != nil {
		return err
	}
	if len(menuPermissions) == 0 {
		return nil
	}

	menuSet := make(map[uuid.UUID]struct{})
	permSet := make(map[uuid.UUID]struct{})
	rows := make([]roledomain.RoleMenuPermission, 0)
	for _, mp := range menuPermissions {
		if mp.MenuID == uuid.Nil {
			return gorm.ErrRecordNotFound
		}
		if len(mp.PermissionIDs) == 0 {
			return gorm.ErrRecordNotFound
		}
		menuSet[mp.MenuID] = struct{}{}
		for _, permissionID := range mp.PermissionIDs {
			if permissionID == uuid.Nil {
				return gorm.ErrRecordNotFound
			}
			permSet[permissionID] = struct{}{}
			rows = append(rows, roledomain.RoleMenuPermission{RoleID: roleID, MenuID: mp.MenuID, PermissionID: permissionID})
		}
	}

	menuIDs := make([]uuid.UUID, 0, len(menuSet))
	for id := range menuSet {
		menuIDs = append(menuIDs, id)
	}
	permissionIDs := make([]uuid.UUID, 0, len(permSet))
	for id := range permSet {
		permissionIDs = append(permissionIDs, id)
	}

	var menuCount int64
	if err := tx.Model(&menudomain.Menu{}).Where("id IN ?", menuIDs).Count(&menuCount).Error; err != nil {
		return err
	}
	if int(menuCount) != len(menuIDs) {
		return gorm.ErrRecordNotFound
	}

	var permissionCount int64
	if err := tx.Model(&permissiondomain.Permission{}).Where("id IN ?", permissionIDs).Count(&permissionCount).Error; err != nil {
		return err
	}
	if int(permissionCount) != len(permissionIDs) {
		return gorm.ErrRecordNotFound
	}

	if err := tx.Create(&rows).Error; err != nil {
		return err
	}
	return nil
}
