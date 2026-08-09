package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	aldomain "apigofiberhorpug/internal/feature/activitylog/domain"
	alusecase "apigofiberhorpug/internal/feature/activitylog/usecase"
	"apigofiberhorpug/internal/feature/bill/domain"
	"apigofiberhorpug/internal/feature/bill/usecase"

	"github.com/gofiber/fiber/v3"
)

type BillHandler struct {
	bills       *usecase.BillUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewBillHandler(bills *usecase.BillUseCase, activityLog *alusecase.ActivityLogUseCase) *BillHandler {
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
	if err := validateCreateBillRequest(&req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	d, err := h.bills.Create(c.Context(), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityCreate, "bill", d.ID, d)
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
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityUpdate, "bill", d.ID, d)
	return response.OK(c, d)
}

func (h *BillHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.bills.Delete(c.Context(), id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityDelete, "bill", id, nil)
	return response.Message(c, "bill deleted")
}
