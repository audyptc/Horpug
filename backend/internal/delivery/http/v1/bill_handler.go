package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"
	"apigofiberhorpug/internal/validator"

	"github.com/gofiber/fiber/v3"
)

type BillHandler struct {
	bills       *usecase.BillUseCase
	activityLog *usecase.ActivityLogUseCase
}

func NewBillHandler(bills *usecase.BillUseCase, activityLog *usecase.ActivityLogUseCase) *BillHandler {
	return &BillHandler{bills: bills, activityLog: activityLog}
}

func (h *BillHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
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
	actorID, _ := c.Locals("user_id").(string)
	d, err := h.bills.Create(c.Context(), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.Log(c.Context(), actorID, domain.ActivityCreate, "bill", d.ID, d)
	return response.Created(c, d)
}

func (h *BillHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateBillRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	actorID, _ := c.Locals("user_id").(string)
	d, err := h.bills.Update(c.Context(), c.Params("id"), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.Log(c.Context(), actorID, domain.ActivityUpdate, "bill", d.ID, d)
	return response.OK(c, d)
}

func (h *BillHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.bills.Delete(c.Context(), id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, domain.ActivityDelete, "bill", id, nil)
	return response.Message(c, "bill deleted")
}
