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

type MeterReadingHandler struct {
	meters *usecase.MeterReadingUseCase
}

func NewMeterReadingHandler(meters *usecase.MeterReadingUseCase) *MeterReadingHandler {
	return &MeterReadingHandler{meters: meters}
}

func (h *MeterReadingHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	list, total, err := h.meters.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *MeterReadingHandler) GetByID(c fiber.Ctx) error {
	d, err := h.meters.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *MeterReadingHandler) Create(c fiber.Ctx) error {
	var req domain.CreateMeterReadingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreateMeterReadingRequest(&req); err != nil {
		return err
	}
	d, err := h.meters.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	return response.Created(c, d)
}

func (h *MeterReadingHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateMeterReadingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	d, err := h.meters.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *MeterReadingHandler) Delete(c fiber.Ctx) error {
	if err := h.meters.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "meter reading deleted")
}
