package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	aldomain "apigofiberhorpug/internal/feature/activitylog/domain"
	alusecase "apigofiberhorpug/internal/feature/activitylog/usecase"
	"apigofiberhorpug/internal/feature/payment/domain"
	"apigofiberhorpug/internal/feature/payment/usecase"

	"github.com/gofiber/fiber/v3"
)

type PaymentHandler struct {
	payment     *usecase.PaymentUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewPaymentHandler(payment *usecase.PaymentUseCase, activityLog *alusecase.ActivityLogUseCase) *PaymentHandler {
	return &PaymentHandler{payment: payment, activityLog: activityLog}
}

func (h *PaymentHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	list, total, err := h.payment.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *PaymentHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.payment.GetByID(c.Context(), dormitoryID, c.Params("id"))
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
	if err := validateCreatePaymentRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.payment.Create(c.Context(), dormitoryID, &req)
	if err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityCreate, "payment", p.ID, p)
	return response.Created(c, p)
}

func (h *PaymentHandler) Update(c fiber.Ctx) error {
	var req domain.UpdatePaymentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateUpdatePaymentRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.payment.Update(c.Context(), dormitoryID, c.Params("id"), &req)
	if err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityUpdate, "payment", p.ID, p)
	return response.OK(c, p)
}

func (h *PaymentHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.payment.Delete(c.Context(), dormitoryID, id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityDelete, "payment", id, nil)
	return response.Message(c, "payment deleted")
}
