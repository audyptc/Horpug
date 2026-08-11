package http

import (
	"context"
	"errors"
	"strings"
	"time"

	dormdomain "apihorpug/internal/features/dormitory/domain"
	dormusecase "apihorpug/internal/features/dormitory/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *dormusecase.Service
}

type createDormitoryRequest struct {
	Name        string      `json:"name"`
	Address     string      `json:"address"`
	Phone       string      `json:"phone"`
	Description string      `json:"description"`
	IsActive    *bool       `json:"is_active"`
	ManagerIDs  []uuid.UUID `json:"manager_ids"`
	CreatedBy   *uuid.UUID  `json:"created_by"`
}

type updateDormitoryRequest struct {
	Name        *string      `json:"name"`
	Address     *string      `json:"address"`
	Phone       *string      `json:"phone"`
	Description *string      `json:"description"`
	IsActive    *bool        `json:"is_active"`
	ManagerIDs  *[]uuid.UUID `json:"manager_ids"`
	UpdatedBy   *uuid.UUID   `json:"updated_by"`
}

func NewHandler(usecase *dormusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

// List godoc
// @Summary List dormitories
// @Tags dormitories
// @Produce json
// @Param user_id query string false "Filter by managing user ID"
// @Success 200 {array} dormdomain.Dormitory
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Router /dormitories [get]
func (h *Handler) List(c fiber.Ctx) error {
	var filter dormusecase.ListFilter
	if raw := c.Query("user_id"); raw != "" {
		userID, err := uuid.Parse(raw)
		if err != nil {
			return apierror.BadRequest("invalid user_id")
		}
		filter.UserID = &userID
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	dormitories, err := h.usecase.List(ctx, filter)
	if err != nil {
		return apierror.Internal("failed to list dormitories")
	}

	return apiresponse.OK(c, dormitories)
}

// Get godoc
// @Summary Get a dormitory by ID
// @Tags dormitories
// @Produce json
// @Param id path string true "Dormitory ID"
// @Success 200 {object} dormdomain.Dormitory
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Router /dormitories/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid dormitory id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	dormitory, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, dormdomain.ErrDormitoryNotFound) {
			return apierror.NotFound("dormitory not found")
		}
		return apierror.Internal("failed to get dormitory")
	}

	return apiresponse.OK(c, dormitory)
}

// Create godoc
// @Summary Create a dormitory
// @Tags dormitories
// @Accept json
// @Produce json
// @Param request body createDormitoryRequest true "Dormitory payload"
// @Success 201 {object} dormdomain.Dormitory
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Router /dormitories [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createDormitoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	dormitory, err := h.usecase.Create(ctx, dormusecase.CreateInput{
		Name:        strings.TrimSpace(req.Name),
		Address:     req.Address,
		Phone:       req.Phone,
		Description: req.Description,
		IsActive:    isActive,
		ManagerIDs:  req.ManagerIDs,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		if errors.Is(err, dormdomain.ErrRequiredDormitoryData) {
			return apierror.BadRequest("name is required")
		}
		if errors.Is(err, dormdomain.ErrManagerNotFound) {
			return apierror.BadRequest("one or more users not found")
		}
		return apierror.Internal("failed to create dormitory")
	}

	return apiresponse.Created(c, dormitory)
}

// Update godoc
// @Summary Update a dormitory
// @Tags dormitories
// @Accept json
// @Produce json
// @Param id path string true "Dormitory ID"
// @Param request body updateDormitoryRequest true "Dormitory payload"
// @Success 200 {object} dormdomain.Dormitory
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Router /dormitories/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid dormitory id")
	}

	var req updateDormitoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	dormitory, err := h.usecase.Update(ctx, id, dormusecase.UpdateInput{
		Name:        req.Name,
		Address:     req.Address,
		Phone:       req.Phone,
		Description: req.Description,
		IsActive:    req.IsActive,
		ManagerIDs:  req.ManagerIDs,
		UpdatedBy:   req.UpdatedBy,
	})
	if err != nil {
		if errors.Is(err, dormdomain.ErrDormitoryNotFound) {
			return apierror.NotFound("dormitory not found")
		}
		if errors.Is(err, dormdomain.ErrRequiredDormitoryData) {
			return apierror.BadRequest("name cannot be empty")
		}
		if errors.Is(err, dormdomain.ErrManagerNotFound) {
			return apierror.BadRequest("one or more users not found")
		}
		return apierror.Internal("failed to update dormitory")
	}

	return apiresponse.OK(c, dormitory)
}

// Delete godoc
// @Summary Delete a dormitory
// @Tags dormitories
// @Produce json
// @Param id path string true "Dormitory ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Router /dormitories/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid dormitory id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id); err != nil {
		if errors.Is(err, dormdomain.ErrDormitoryNotFound) {
			return apierror.NotFound("dormitory not found")
		}
		return apierror.Internal("failed to delete dormitory")
	}

	return apiresponse.Message(c, "dormitory deleted")
}
