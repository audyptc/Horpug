package http

import (
	"context"
	"errors"
	"strings"
	"time"

	userdomain "apihorpug/internal/features/user/domain"
	userusecase "apihorpug/internal/features/user/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *userusecase.Service
}

type createUserRequest struct {
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	RoleID    uuid.UUID  `json:"role_id"`
	IsActive  *bool      `json:"is_active"`
	CreatedBy *uuid.UUID `json:"created_by"`
}

type updateUserRequest struct {
	Username  *string    `json:"username"`
	Email     *string    `json:"email"`
	Password  *string    `json:"password"`
	RoleID    *uuid.UUID `json:"role_id"`
	IsActive  *bool      `json:"is_active"`
	UpdatedBy *uuid.UUID `json:"updated_by"`
}

func NewHandler(usecase *userusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

// List godoc
// @Summary List users
// @Tags users
// @Produce json
// @Success 200 {array} userdomain.User
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /users [get]
func (h *Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	users, err := h.usecase.List(ctx)
	if err != nil {
		return apierror.Internal("failed to list users")
	}

	return apiresponse.OK(c, users)
}

// Get godoc
// @Summary Get a user by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} userdomain.User
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid user id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	user, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return apierror.NotFound("user not found")
		}
		return apierror.Internal("failed to get user")
	}

	return apiresponse.OK(c, user)
}

// GetPermissions godoc
// @Summary Get a user's permissions
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {array} userusecase.UserPermissionItem
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /users/{id}/permissions [get]
func (h *Handler) GetPermissions(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid user id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	permissions, err := h.usecase.GetPermissions(ctx, id)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return apierror.NotFound("user not found")
		}
		return apierror.Internal("failed to load user permissions")
	}

	return apiresponse.OK(c, permissions)
}

// Create godoc
// @Summary Create a user
// @Tags users
// @Accept json
// @Produce json
// @Param request body createUserRequest true "User payload"
// @Success 201 {object} userdomain.User
// @Failure 400 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /users [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	user, err := h.usecase.Create(ctx, userusecase.CreateInput{
		Username:  strings.TrimSpace(req.Username),
		Email:     strings.TrimSpace(req.Email),
		Password:  req.Password,
		RoleID:    req.RoleID,
		IsActive:  isActive,
		CreatedBy: req.CreatedBy,
	})
	if err != nil {
		if errors.Is(err, userdomain.ErrRequiredUserData) {
			return apierror.BadRequest("username, email and password are required")
		}
		if errors.Is(err, userdomain.ErrRoleNotFound) {
			if req.RoleID == uuid.Nil {
				return apierror.BadRequest("role_id is required")
			}
			return apierror.BadRequest("role not found")
		}
		if errors.Is(err, userdomain.ErrUserDuplicate) {
			return apierror.Conflict("username or email already exists")
		}
		return apierror.Internal("failed to create user")
	}

	return apiresponse.Created(c, user)
}

// Update godoc
// @Summary Update a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body updateUserRequest true "User payload"
// @Success 200 {object} userdomain.User
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /users/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid user id")
	}

	var req updateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	updatedUser, err := h.usecase.Update(ctx, id, userusecase.UpdateInput{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		RoleID:    req.RoleID,
		IsActive:  req.IsActive,
		UpdatedBy: req.UpdatedBy,
	})
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return apierror.NotFound("user not found")
		}
		if errors.Is(err, userdomain.ErrInvalidUsername) {
			return apierror.BadRequest("username cannot be empty")
		}
		if errors.Is(err, userdomain.ErrInvalidEmail) {
			return apierror.BadRequest("email cannot be empty")
		}
		if errors.Is(err, userdomain.ErrInvalidPassword) {
			return apierror.BadRequest("password cannot be empty")
		}
		if errors.Is(err, userdomain.ErrRoleNotFound) {
			return apierror.BadRequest("role not found")
		}
		if errors.Is(err, userdomain.ErrUserDuplicate) {
			return apierror.Conflict("username or email already exists")
		}
		return apierror.Internal("failed to update user")
	}

	return apiresponse.OK(c, updatedUser)
}

// Delete godoc
// @Summary Delete a user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /users/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid user id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id); err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return apierror.NotFound("user not found")
		}
		return apierror.Internal("failed to delete user")
	}

	return apiresponse.Message(c, "user deleted")
}
