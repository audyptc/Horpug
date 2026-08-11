package http

import (
	"errors"
	"strings"

	"apihorpug/internal/models"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Server struct {
	db *gorm.DB
}

type createRoleRequest struct {
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	IsActive      *bool       `json:"is_active"`
	PermissionIDs []uuid.UUID `json:"permission_ids"`
}

type updateRoleRequest struct {
	Name          *string      `json:"name"`
	Description   *string      `json:"description"`
	IsActive      *bool        `json:"is_active"`
	PermissionIDs *[]uuid.UUID `json:"permission_ids"`
}

type createUserRequest struct {
	Username string     `json:"username"`
	Email    string     `json:"email"`
	Password string     `json:"password"`
	RoleID   *uuid.UUID `json:"role_id"`
	IsActive *bool      `json:"is_active"`
}

type updateUserRequest struct {
	Username *string    `json:"username"`
	Email    *string    `json:"email"`
	Password *string    `json:"password"`
	RoleID   *uuid.UUID `json:"role_id"`
	IsActive *bool      `json:"is_active"`
}

type createPermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	s := &Server{db: db}

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1")

	api.Get("/permissions", s.ListPermissions)
	api.Post("/permissions", s.CreatePermission)

	api.Get("/roles", s.ListRoles)
	api.Get("/roles/:id", s.GetRole)
	api.Post("/roles", s.CreateRole)
	api.Put("/roles/:id", s.UpdateRole)
	api.Delete("/roles/:id", s.DeleteRole)

	api.Get("/users", s.ListUsers)
	api.Get("/users/:id", s.GetUser)
	api.Get("/users/:id/permissions", s.GetUserPermissions)
	api.Post("/users", s.CreateUser)
	api.Put("/users/:id", s.UpdateUser)
	api.Delete("/users/:id", s.DeleteUser)
}

func (s *Server) ListPermissions(c fiber.Ctx) error {
	var permissions []models.Permission
	if err := s.db.Order("name asc").Find(&permissions).Error; err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "failed to list permissions")
	}
	return c.JSON(permissions)
}

func (s *Server) CreatePermission(c fiber.Ctx) error {
	var req createPermissionRequest
	if err := c.Bind().Body(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid request body")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return errorJSON(c, fiber.StatusBadRequest, "name is required")
	}

	permission := models.Permission{
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
	}

	if err := s.db.Create(&permission).Error; err != nil {
		if isUniqueViolation(err) {
			return errorJSON(c, fiber.StatusConflict, "permission name already exists")
		}
		return errorJSON(c, fiber.StatusInternalServerError, "failed to create permission")
	}

	return c.Status(fiber.StatusCreated).JSON(permission)
}

func (s *Server) ListRoles(c fiber.Ctx) error {
	var roles []models.Role
	if err := s.db.Preload("Permissions").Order("name asc").Find(&roles).Error; err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "failed to list roles")
	}
	return c.JSON(roles)
}

func (s *Server) GetRole(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid role id")
	}

	var role models.Role
	if err := s.db.Preload("Permissions").First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorJSON(c, fiber.StatusNotFound, "role not found")
		}
		return errorJSON(c, fiber.StatusInternalServerError, "failed to get role")
	}

	return c.JSON(role)
}

func (s *Server) CreateRole(c fiber.Ctx) error {
	var req createRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid request body")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return errorJSON(c, fiber.StatusBadRequest, "name is required")
	}

	role := models.Role{
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
		IsActive:    true,
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		return replaceRolePermissions(tx, role.ID, req.PermissionIDs)
	})
	if err != nil {
		if isUniqueViolation(err) {
			return errorJSON(c, fiber.StatusConflict, "role name already exists")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorJSON(c, fiber.StatusBadRequest, "one or more permissions not found")
		}
		return errorJSON(c, fiber.StatusInternalServerError, "failed to create role")
	}

	if err := s.db.Preload("Permissions").First(&role, "id = ?", role.ID).Error; err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "failed to load role")
	}

	return c.Status(fiber.StatusCreated).JSON(role)
}

func (s *Server) UpdateRole(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid role id")
	}

	var req updateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid request body")
	}

	var role models.Role
	if err := s.db.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorJSON(c, fiber.StatusNotFound, "role not found")
		}
		return errorJSON(c, fiber.StatusInternalServerError, "failed to get role")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
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

		if req.PermissionIDs != nil {
			if err := replaceRolePermissions(tx, role.ID, *req.PermissionIDs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return errorJSON(c, fiberErr.Code, fiberErr.Message)
		}
		if isUniqueViolation(err) {
			return errorJSON(c, fiber.StatusConflict, "role name already exists")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorJSON(c, fiber.StatusBadRequest, "one or more permissions not found")
		}
		return errorJSON(c, fiber.StatusInternalServerError, "failed to update role")
	}

	if err := s.db.Preload("Permissions").First(&role, "id = ?", role.ID).Error; err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "failed to load role")
	}

	return c.JSON(role)
}

func (s *Server) DeleteRole(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid role id")
	}

	result := s.db.Delete(&models.Role{}, "id = ?", id)
	if result.Error != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "failed to delete role")
	}
	if result.RowsAffected == 0 {
		return errorJSON(c, fiber.StatusNotFound, "role not found")
	}

	return c.JSON(fiber.Map{"message": "role deleted"})
}

func (s *Server) ListUsers(c fiber.Ctx) error {
	var users []models.User
	if err := s.db.Preload("Role").Order("created_at desc").Find(&users).Error; err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "failed to list users")
	}
	return c.JSON(users)
}

func (s *Server) GetUser(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid user id")
	}

	var user models.User
	if err := s.db.Preload("Role").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorJSON(c, fiber.StatusNotFound, "user not found")
		}
		return errorJSON(c, fiber.StatusInternalServerError, "failed to get user")
	}

	return c.JSON(user)
}

func (s *Server) GetUserPermissions(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid user id")
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorJSON(c, fiber.StatusNotFound, "user not found")
		}
		return errorJSON(c, fiber.StatusInternalServerError, "failed to get user")
	}

	if user.RoleID == nil {
		return c.JSON([]models.Permission{})
	}

	var permissions []models.Permission
	err = s.db.
		Table("permissions").
		Select("permissions.*").
		Joins("join role_permissions on role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", *user.RoleID).
		Order("permissions.name asc").
		Find(&permissions).Error
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "failed to load user permissions")
	}

	return c.JSON(permissions)
}

func (s *Server) CreateUser(c fiber.Ctx) error {
	var req createUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid request body")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return errorJSON(c, fiber.StatusBadRequest, "username, email and password are required")
	}

	if req.RoleID != nil {
		var count int64
		if err := s.db.Model(&models.Role{}).Where("id = ?", *req.RoleID).Count(&count).Error; err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, "failed to validate role")
		}
		if count == 0 {
			return errorJSON(c, fiber.StatusBadRequest, "role not found")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "failed to hash password")
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		RoleID:   req.RoleID,
		IsActive: true,
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.db.Create(&user).Error; err != nil {
		if isUniqueViolation(err) {
			return errorJSON(c, fiber.StatusConflict, "username or email already exists")
		}
		return errorJSON(c, fiber.StatusInternalServerError, "failed to create user")
	}

	if err := s.db.Preload("Role").First(&user, "id = ?", user.ID).Error; err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "failed to load user")
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (s *Server) UpdateUser(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid user id")
	}

	var req updateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid request body")
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorJSON(c, fiber.StatusNotFound, "user not found")
		}
		return errorJSON(c, fiber.StatusInternalServerError, "failed to get user")
	}

	updates := map[string]any{}
	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if username == "" {
			return errorJSON(c, fiber.StatusBadRequest, "username cannot be empty")
		}
		updates["username"] = username
	}
	if req.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*req.Email))
		if email == "" {
			return errorJSON(c, fiber.StatusBadRequest, "email cannot be empty")
		}
		updates["email"] = email
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.RoleID != nil {
		var count int64
		if err := s.db.Model(&models.Role{}).Where("id = ?", *req.RoleID).Count(&count).Error; err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, "failed to validate role")
		}
		if count == 0 {
			return errorJSON(c, fiber.StatusBadRequest, "role not found")
		}
		updates["role_id"] = *req.RoleID
	}
	if req.Password != nil {
		if strings.TrimSpace(*req.Password) == "" {
			return errorJSON(c, fiber.StatusBadRequest, "password cannot be empty")
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, "failed to hash password")
		}
		updates["password"] = string(hashedPassword)
	}

	if len(updates) > 0 {
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			if isUniqueViolation(err) {
				return errorJSON(c, fiber.StatusConflict, "username or email already exists")
			}
			return errorJSON(c, fiber.StatusInternalServerError, "failed to update user")
		}
	}

	if err := s.db.Preload("Role").First(&user, "id = ?", user.ID).Error; err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "failed to load user")
	}

	return c.JSON(user)
}

func (s *Server) DeleteUser(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid user id")
	}

	result := s.db.Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "failed to delete user")
	}
	if result.RowsAffected == 0 {
		return errorJSON(c, fiber.StatusNotFound, "user not found")
	}

	return c.JSON(fiber.Map{"message": "user deleted"})
}

func replaceRolePermissions(tx *gorm.DB, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
		return err
	}
	if len(permissionIDs) == 0 {
		return nil
	}

	var permissions []models.Permission
	if err := tx.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		return err
	}
	if len(permissions) != len(permissionIDs) {
		return gorm.ErrRecordNotFound
	}

	items := make([]models.RolePermission, 0, len(permissionIDs))
	for _, permissionID := range permissionIDs {
		items = append(items, models.RolePermission{RoleID: roleID, PermissionID: permissionID})
	}
	return tx.Create(&items).Error
}

func errorJSON(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": message,
	})
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}
