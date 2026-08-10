package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/expense/domain"
	"apigofiberhorpug/internal/feature/expense/usecase"

	"github.com/gofiber/fiber/v3"
)

type ExpenseHandler struct {
	expenses *usecase.ExpenseUseCase
}

func NewExpenseHandler(expenses *usecase.ExpenseUseCase) *ExpenseHandler {
	return &ExpenseHandler{expenses: expenses}
}

func (h *ExpenseHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	list, total, err := h.expenses.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *ExpenseHandler) GetByID(c fiber.Ctx) error {
	e, err := h.expenses.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, e)
}

func (h *ExpenseHandler) Create(c fiber.Ctx) error {
	var req domain.CreateExpenseRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateExpenseRequest(&req); err != nil {
		return err
	}
	e, err := h.expenses.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	return response.Created(c, e)
}

func (h *ExpenseHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateExpenseRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateUpdateExpenseRequest(&req); err != nil {
		return err
	}
	e, err := h.expenses.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, e)
}

func (h *ExpenseHandler) Delete(c fiber.Ctx) error {
	if err := h.expenses.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "expense deleted")
}
