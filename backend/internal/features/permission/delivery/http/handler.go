package http

import (
	"context"
	"errors"
	"time"

	permissiondomain "apihorpug/internal/features/permission/domain"
	permissionusecase "apihorpug/internal/features/permission/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	usecase *permissionusecase.Service
}

type createPermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewHandler(usecase *permissionusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

// List godoc
// @Summary List permissions
// @Tags permissions
// @Produce json
// @Success 200 {array} permissiondomain.Permission
// @Failure 500 {object} apierror.Error
// @Router /permissions [get]
func (h *Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	permissions, err := h.usecase.List(ctx)
	if err != nil {
		return apierror.Internal("failed to list permissions")
	}

	return apiresponse.OK(c, permissions)
}

// Create godoc
// @Summary Create a permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param request body createPermissionRequest true "Permission payload"
// @Success 201 {object} permissiondomain.Permission
// @Failure 400 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Router /permissions [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createPermissionRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	permission, err := h.usecase.Create(ctx, permissionusecase.CreateInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, permissiondomain.ErrPermissionNameRequired) {
			return apierror.BadRequest("name is required")
		}
		if errors.Is(err, permissiondomain.ErrPermissionNameInvalid) {
			return apierror.BadRequest("name must be one of the predefined actions")
		}
		if errors.Is(err, permissiondomain.ErrPermissionNameExists) {
			return apierror.Conflict("permission name already exists")
		}
		return apierror.Internal("failed to create permission")
	}

	return apiresponse.Created(c, permission)
}
