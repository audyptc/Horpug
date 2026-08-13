package http

import (
	"context"
	"errors"
	"strings"
	"time"

	expensedomain "apihorpug/internal/features/expense/domain"
	expenseusecase "apihorpug/internal/features/expense/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *expenseusecase.Service
}

type createExpenseRequest struct {
	DormitoryID uuid.UUID                     `json:"dormitory_id"`
	Category    expensedomain.ExpenseCategory `json:"category"`
	ExpenseDate time.Time                     `json:"expense_date"`
	Amount      float64                       `json:"amount"`
	Description string                        `json:"description"`
}

type updateExpenseRequest struct {
	Category    *expensedomain.ExpenseCategory `json:"category"`
	ExpenseDate *time.Time                     `json:"expense_date"`
	Amount      *float64                       `json:"amount"`
	Description *string                        `json:"description"`
}

func NewHandler(usecase *expenseusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

func parseUUIDQuery(c fiber.Ctx, name string) (*uuid.UUID, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, apierror.BadRequest("invalid " + name)
	}
	return &id, nil
}

func parseDateQuery(c fiber.Ctx, name string) (*time.Time, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	value, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, apierror.BadRequest("invalid " + name)
	}
	return &value, nil
}

func parseListFilters(c fiber.Ctx) (expenseusecase.ListFilters, error) {
	dormitoryID, err := parseUUIDQuery(c, "dormitory_id")
	if err != nil {
		return expenseusecase.ListFilters{}, err
	}
	dateFrom, err := parseDateQuery(c, "date_from")
	if err != nil {
		return expenseusecase.ListFilters{}, err
	}
	dateTo, err := parseDateQuery(c, "date_to")
	if err != nil {
		return expenseusecase.ListFilters{}, err
	}

	var category *expensedomain.ExpenseCategory
	if raw := strings.TrimSpace(c.Query("category")); raw != "" {
		cat := expensedomain.ExpenseCategory(raw)
		if !cat.Valid() {
			return expenseusecase.ListFilters{}, apierror.BadRequest("invalid category")
		}
		category = &cat
	}

	return expenseusecase.ListFilters{
		DormitoryID: dormitoryID,
		Category:    category,
		DateFrom:    dateFrom,
		DateTo:      dateTo,
	}, nil
}

// List godoc
// @Summary List expenses
// @Description Returns every expense for roles with full dormitory access, otherwise only expenses under dormitories the caller manages. Optionally filter by dormitory, category or expense date range.
// @Tags expenses
// @Produce json
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param category query string false "Filter by category (maintenance, utility, salary, supplies, other)"
// @Param date_from query string false "Filter by expense date, inclusive (YYYY-MM-DD)"
// @Param date_to query string false "Filter by expense date, inclusive (YYYY-MM-DD)"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /expenses [get]
func (h *Handler) List(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

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

	expenses, total, err := h.usecase.List(ctx, requesterID, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list expenses")
	}

	return apiresponse.Paginated(c, expenses, page, perPage, total)
}

// Get godoc
// @Summary Get an expense by ID
// @Tags expenses
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} expensedomain.Expense
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /expenses/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid expense id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	expense, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, expensedomain.ErrExpenseNotFound) {
			return apierror.NotFound("expense not found")
		}
		return apierror.Internal("failed to get expense")
	}

	return apiresponse.OK(c, expense)
}

// Create godoc
// @Summary Create an expense
// @Description Records an operating expense (e.g. maintenance, salary, supplies) against a dormitory.
// @Tags expenses
// @Accept json
// @Produce json
// @Param request body createExpenseRequest true "Expense payload"
// @Success 201 {object} expensedomain.Expense
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /expenses [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createExpenseRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	expense, err := h.usecase.Create(ctx, expenseusecase.CreateInput{
		DormitoryID: req.DormitoryID,
		Category:    req.Category,
		ExpenseDate: req.ExpenseDate,
		Amount:      req.Amount,
		Description: req.Description,
		CreatedBy:   &requesterID,
	})
	if err != nil {
		if errors.Is(err, expensedomain.ErrRequiredExpenseData) {
			return apierror.BadRequest("dormitory_id and expense_date are required")
		}
		if errors.Is(err, expensedomain.ErrInvalidExpenseAmount) {
			return apierror.BadRequest("amount must be greater than zero")
		}
		if errors.Is(err, expensedomain.ErrInvalidCategory) {
			return apierror.BadRequest("invalid expense category")
		}
		if errors.Is(err, expensedomain.ErrDormitoryNotFound) {
			return apierror.NotFound("dormitory not found")
		}
		return apierror.Internal("failed to create expense")
	}

	return apiresponse.Created(c, expense)
}

// Update godoc
// @Summary Update an expense
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Param request body updateExpenseRequest true "Expense payload"
// @Success 200 {object} expensedomain.Expense
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /expenses/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid expense id")
	}

	var req updateExpenseRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	expense, err := h.usecase.Update(ctx, id, requesterID, expenseusecase.UpdateInput{
		Category:    req.Category,
		ExpenseDate: req.ExpenseDate,
		Amount:      req.Amount,
		Description: req.Description,
		UpdatedBy:   &requesterID,
	})
	if err != nil {
		if errors.Is(err, expensedomain.ErrExpenseNotFound) {
			return apierror.NotFound("expense not found")
		}
		if errors.Is(err, expensedomain.ErrInvalidExpenseAmount) {
			return apierror.BadRequest("amount must be greater than zero")
		}
		if errors.Is(err, expensedomain.ErrInvalidCategory) {
			return apierror.BadRequest("invalid expense category")
		}
		return apierror.Internal("failed to update expense")
	}

	return apiresponse.OK(c, expense)
}

// Delete godoc
// @Summary Delete an expense
// @Tags expenses
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /expenses/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid expense id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID); err != nil {
		if errors.Is(err, expensedomain.ErrExpenseNotFound) {
			return apierror.NotFound("expense not found")
		}
		return apierror.Internal("failed to delete expense")
	}

	return apiresponse.Message(c, "expense deleted")
}
