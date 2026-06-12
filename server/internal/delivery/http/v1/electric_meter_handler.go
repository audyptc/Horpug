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

type ElectricMeterHandler struct {
	uc *usecase.ElectricMeterUseCase
}

func NewElectricMeterHandler(uc *usecase.ElectricMeterUseCase) *ElectricMeterHandler {
	return &ElectricMeterHandler{uc: uc}
}

func (h *ElectricMeterHandler) List(c fiber.Ctx) error {
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

func (h *ElectricMeterHandler) GetByID(c fiber.Ctx) error {
	d, err := h.uc.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *ElectricMeterHandler) Create(c fiber.Ctx) error {
	var req domain.CreateElectricMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreateElectricMeterRequest(&req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	d, err := h.uc.Create(c.Context(), &req, actorID)
	if err != nil {
		return err
	}
	return response.Created(c, d)
}

func (h *ElectricMeterHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateElectricMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	actorID, _ := c.Locals("user_id").(string)
	d, err := h.uc.Update(c.Context(), c.Params("id"), &req, actorID)
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *ElectricMeterHandler) Delete(c fiber.Ctx) error {
	if err := h.uc.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "electric meter reading deleted")
}
