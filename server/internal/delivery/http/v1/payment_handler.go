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

type PaymentHandler struct {
	payment     *usecase.PaymentUseCase
	activityLog *usecase.ActivityLogUseCase
}

func NewPaymentHandler(payment *usecase.PaymentUseCase, activityLog *usecase.ActivityLogUseCase) *PaymentHandler {
	return &PaymentHandler{payment: payment, activityLog: activityLog}
}

func (h *PaymentHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	list, total, err := h.payment.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *PaymentHandler) GetByID(c fiber.Ctx) error {
	p, err := h.payment.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

func (h *PaymentHandler) Create(c fiber.Ctx) error {
	var req domain.CreatePaymentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreatePaymentRequest(&req); err != nil {
		return err
	}
	p, err := h.payment.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, domain.ActivityCreate, "payment", p.ID, p)
	return response.Created(c, p)
}

func (h *PaymentHandler) Update(c fiber.Ctx) error {
	var req domain.UpdatePaymentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.UpdatePaymentRequest(&req); err != nil {
		return err
	}
	p, err := h.payment.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, domain.ActivityUpdate, "payment", p.ID, p)
	return response.OK(c, p)
}

func (h *PaymentHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.payment.Delete(c.Context(), id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, domain.ActivityDelete, "payment", id, nil)
	return response.Message(c, "payment deleted")
}
