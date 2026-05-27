package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"
	"apigofiberhorpug/internal/validator"

	"github.com/gofiber/fiber/v3"
)

type ParcelHandler struct {
	parcels *usecase.ParcelUseCase
}

func NewParcelHandler(parcels *usecase.ParcelUseCase) *ParcelHandler {
	return &ParcelHandler{parcels: parcels}
}

func (h *ParcelHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := parsePaginationQuery(c)
	if err != nil {
		return err
	}
	list, total, err := h.parcels.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *ParcelHandler) GetByID(c fiber.Ctx) error {
	p, err := h.parcels.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

func (h *ParcelHandler) Create(c fiber.Ctx) error {
	var req domain.CreateParcelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreateParcelRequest(&req); err != nil {
		return err
	}
	p, err := h.parcels.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	return response.Created(c, p)
}

func (h *ParcelHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateParcelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.UpdateParcelRequest(&req); err != nil {
		return err
	}
	p, err := h.parcels.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

func (h *ParcelHandler) Delete(c fiber.Ctx) error {
	if err := h.parcels.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "parcel deleted")
}
