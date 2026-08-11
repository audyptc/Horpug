package http

import (
	"context"
	"errors"
	"strings"
	"time"

	userdomain "apihorpug/internal/features/user/domain"
	userusecase "apihorpug/internal/features/user/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *userusecase.Service
}

type createUserRequest struct {
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	RoleID   uuid.UUID `json:"role_id"`
	IsActive *bool     `json:"is_active"`
}

type updateUserRequest struct {
	Username *string    `json:"username"`
	Email    *string    `json:"email"`
	Password *string    `json:"password"`
	RoleID   *uuid.UUID `json:"role_id"`
	IsActive *bool      `json:"is_active"`
}

func NewHandler(usecase *userusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	users, err := h.usecase.List(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list users"})
	}

	return c.JSON(users)
}

func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	user, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get user"})
	}

	return c.JSON(user)
}

func (h *Handler) GetPermissions(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	permissions, err := h.usecase.GetPermissions(ctx, id)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load user permissions"})
	}

	return c.JSON(permissions)
}

func (h *Handler) Create(c fiber.Ctx) error {
	var req createUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	user, err := h.usecase.Create(ctx, userusecase.CreateInput{
		Username: strings.TrimSpace(req.Username),
		Email:    strings.TrimSpace(req.Email),
		Password: req.Password,
		RoleID:   req.RoleID,
		IsActive: isActive,
	})
	if err != nil {
		if errors.Is(err, userdomain.ErrRequiredUserData) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username, email and password are required"})
		}
		if errors.Is(err, userdomain.ErrRoleNotFound) {
			if req.RoleID == uuid.Nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "role_id is required"})
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "role not found"})
		}
		if errors.Is(err, userdomain.ErrUserDuplicate) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "username or email already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	var req updateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	updatedUser, err := h.usecase.Update(ctx, id, userusecase.UpdateInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		RoleID:   req.RoleID,
		IsActive: req.IsActive,
	})
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		if errors.Is(err, userdomain.ErrInvalidUsername) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username cannot be empty"})
		}
		if errors.Is(err, userdomain.ErrInvalidEmail) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email cannot be empty"})
		}
		if errors.Is(err, userdomain.ErrInvalidPassword) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password cannot be empty"})
		}
		if errors.Is(err, userdomain.ErrRoleNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "role not found"})
		}
		if errors.Is(err, userdomain.ErrUserDuplicate) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "username or email already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update user"})
	}

	return c.JSON(updatedUser)
}

func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id); err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete user"})
	}

	return c.JSON(fiber.Map{"message": "user deleted"})
}
