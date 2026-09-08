package http

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	userdomain "apihorpug/internal/features/user/domain"
	userusecase "apihorpug/internal/features/user/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

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

const defaultActiveListLimit = 50

// parseOptionalBool reads a tri-state query flag: absent means "don't filter
// on this", so an unset value is distinct from an explicit false.
func parseOptionalBool(c fiber.Ctx, name string) (*bool, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, apierror.BadRequest(name + " must be true or false")
	}
	return &value, nil
}

// parseColumnFilters reads the per-column filters, passed as f[<column>]=value.
// An unrecognised column is rejected rather than ignored: silently dropping it
// would return an unfiltered list that looks like a legitimate result.
func parseColumnFilters(c fiber.Ctx) (map[string]string, error) {
	columns := make(map[string]string)

	for key, value := range c.Queries() {
		if !strings.HasPrefix(key, "f[") || !strings.HasSuffix(key, "]") {
			continue
		}

		name := key[len("f[") : len(key)-len("]")]
		if _, ok := userusecase.FilterColumns[name]; !ok {
			return nil, apierror.BadRequest("unsupported filter field: " + name)
		}

		if value = strings.TrimSpace(value); value != "" {
			columns[name] = value
		}
	}

	return columns, nil
}

func parseListFilters(c fiber.Ctx) (userusecase.ListFilters, error) {
	isActive, err := parseOptionalBool(c, "is_active")
	if err != nil {
		return userusecase.ListFilters{}, err
	}

	columns, err := parseColumnFilters(c)
	if err != nil {
		return userusecase.ListFilters{}, err
	}

	sortKey := userusecase.DefaultSortKey
	if raw := strings.TrimSpace(c.Query("sort")); raw != "" {
		if _, ok := userusecase.SortColumns[raw]; !ok {
			return userusecase.ListFilters{}, apierror.BadRequest("unsupported sort field")
		}
		sortKey = raw
	}

	// Usernames read most naturally sorted ascending, so unlike lists
	// defaulting to a newest-first field, no key flips the default direction.
	sortDesc := false
	switch strings.TrimSpace(c.Query("order")) {
	case "":
	case "asc":
		sortDesc = false
	case "desc":
		sortDesc = true
	default:
		return userusecase.ListFilters{}, apierror.BadRequest("order must be asc or desc")
	}

	return userusecase.ListFilters{
		Search:   strings.TrimSpace(c.Query("q")),
		Columns:  columns,
		IsActive: isActive,
		SortKey:  sortKey,
		SortDesc: sortDesc,
	}, nil
}

// List godoc
// @Summary List users
// @Tags users
// @Produce json
// @Param q query string false "Filter by username, email or role"
// @Param f[column] query string false "Per-column substring filter, e.g. f[username]=admin; column must be one of username, email, role"
// @Param is_active query bool false "Filter by active status"
// @Param sort query string false "Sort field: username, email, role, is_active, created_at (default username)"
// @Param order query string false "Sort direction: asc or desc"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /users [get]
func (h *Handler) List(c fiber.Ctx) error {
	filters, err := parseListFilters(c)
	if err != nil {
		return err
	}

	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	users, total, err := h.usecase.List(ctx, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list users")
	}

	return apiresponse.Paginated(c, users, page, perPage, total)
}

// ListActive godoc
// @Summary List active users
// @Description Returns active users. Intended for populating user selectors; results are capped rather than paginated.
// @Tags users
// @Produce json
// @Param q query string false "Filter by username or email"
// @Param limit query int false "Max results (default 50, max 100)"
// @Success 200 {array} userdomain.User
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /users/active [get]
func (h *Handler) ListActive(c fiber.Ctx) error {
	search := strings.TrimSpace(c.Query("q"))

	limit := defaultActiveListLimit
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return apierror.BadRequest("limit must be a positive integer")
		}
		limit = parsed
	}
	if limit > httputil.MaxPerPage {
		limit = httputil.MaxPerPage
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	users, err := h.usecase.ListActive(ctx, search, limit)
	if err != nil {
		return apierror.Internal("failed to list active users")
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

// CheckDeletion godoc
// @Summary Check whether a user can be deleted
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} userusecase.DeletionCheck
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /users/{id}/deletion-check [get]
func (h *Handler) CheckDeletion(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid user id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	check, err := h.usecase.CheckDeletion(ctx, id)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return apierror.NotFound("user not found")
		}
		return apierror.Internal("failed to check user deletion")
	}

	return apiresponse.OK(c, check)
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

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	user, err := h.usecase.Create(ctx, userusecase.CreateInput{
		Username:  strings.TrimSpace(req.Username),
		Email:     strings.TrimSpace(req.Email),
		Password:  req.Password,
		RoleID:    req.RoleID,
		IsActive:  isActive,
		CreatedBy: &requesterID,
	}, c.IP())
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

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	updatedUser, err := h.usecase.Update(ctx, id, userusecase.UpdateInput{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		RoleID:    req.RoleID,
		IsActive:  req.IsActive,
		UpdatedBy: &requesterID,
	}, c.IP())
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
		if errors.Is(err, userdomain.ErrUserProtected) {
			return apierror.Forbidden("this user is protected and cannot be modified")
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

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID, c.IP()); err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return apierror.NotFound("user not found")
		}
		if errors.Is(err, userdomain.ErrUserProtected) {
			return apierror.Forbidden("this user is protected and cannot be deleted")
		}
		return apierror.Internal("failed to delete user")
	}

	return apiresponse.Message(c, "user deleted")
}
