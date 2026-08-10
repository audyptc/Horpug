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

// List godoc
// @Summary      รายการค่าใช้จ่าย
// @Tags         expenses
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Expense
// @Router       /expenses [get]
func (h *ExpenseHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	list, total, err := h.expenses.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

// GetByID godoc
// @Summary      ดูข้อมูลค่าใช้จ่ายตาม ID
// @Tags         expenses
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Expense ID"
// @Success      200  {object}  domain.Expense
// @Router       /expenses/{id} [get]
func (h *ExpenseHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	e, err := h.expenses.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, e)
}

// Create godoc
// @Summary      สร้างค่าใช้จ่าย
// @Tags         expenses
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateExpenseRequest true "Expense payload"
// @Success      201  {object}  domain.Expense
// @Router       /expenses [post]
func (h *ExpenseHandler) Create(c fiber.Ctx) error {
	var req domain.CreateExpenseRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateExpenseRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	e, err := h.expenses.Create(c.Context(), dormitoryID, &req)
	if err != nil {
		return err
	}
	return response.Created(c, e)
}

// Update godoc
// @Summary      แก้ไขค่าใช้จ่าย
// @Tags         expenses
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Expense ID"
// @Param        body body domain.UpdateExpenseRequest true "Expense payload"
// @Success      200  {object}  domain.Expense
// @Router       /expenses/{id} [put]
func (h *ExpenseHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateExpenseRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateUpdateExpenseRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	e, err := h.expenses.Update(c.Context(), dormitoryID, c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, e)
}

// Delete godoc
// @Summary      ลบค่าใช้จ่าย
// @Tags         expenses
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Expense ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /expenses/{id} [delete]
func (h *ExpenseHandler) Delete(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.expenses.Delete(c.Context(), dormitoryID, c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "expense deleted")
}
