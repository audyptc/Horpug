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

type WaterMeterHandler struct {
	uc *usecase.WaterMeterUseCase
}

func NewWaterMeterHandler(uc *usecase.WaterMeterUseCase) *WaterMeterHandler {
	return &WaterMeterHandler{uc: uc}
}

func (h *WaterMeterHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	list, total, err := h.uc.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *WaterMeterHandler) GetByID(c fiber.Ctx) error {
	d, err := h.uc.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *WaterMeterHandler) Create(c fiber.Ctx) error {
	var req domain.CreateWaterMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreateWaterMeterRequest(&req); err != nil {
		return err
	}
	d, err := h.uc.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	return response.Created(c, d)
}

func (h *WaterMeterHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateWaterMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	d, err := h.uc.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *WaterMeterHandler) Delete(c fiber.Ctx) error {
	if err := h.uc.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "water meter reading deleted")
}
