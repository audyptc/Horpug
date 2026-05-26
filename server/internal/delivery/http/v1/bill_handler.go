package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"
	"apigofiberhorpug/internal/validator"

	"github.com/gofiber/fiber/v3"
)

type BillHandler struct {
	bills *usecase.BillUseCase
}

func NewBillHandler(bills *usecase.BillUseCase) *BillHandler {
	return &BillHandler{bills: bills}
}

func (h *BillHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := parsePaginationQuery(c)
	if err != nil {
		return err
	}
	list, total, err := h.bills.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *BillHandler) GetByID(c fiber.Ctx) error {
	d, err := h.bills.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *BillHandler) Create(c fiber.Ctx) error {
	var req domain.CreateBillRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreateBillRequest(&req); err != nil {
		return err
	}
	d, err := h.bills.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	return response.Created(c, d)
}

func (h *BillHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateBillRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	d, err := h.bills.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *BillHandler) Delete(c fiber.Ctx) error {
	if err := h.bills.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "bill deleted")
}
