package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"
	"apigofiberhorpug/internal/validator"

	"github.com/gofiber/fiber/v3"
)

type ParkingHandler struct {
	parking *usecase.ParkingUseCase
}

func NewParkingHandler(parking *usecase.ParkingUseCase) *ParkingHandler {
	return &ParkingHandler{parking: parking}
}

func (h *ParkingHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := parsePaginationQuery(c)
	if err != nil {
		return err
	}
	list, total, err := h.parking.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *ParkingHandler) GetByID(c fiber.Ctx) error {
	p, err := h.parking.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

func (h *ParkingHandler) Create(c fiber.Ctx) error {
	var req domain.CreateParkingSlotRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreateParkingSlotRequest(&req); err != nil {
		return err
	}
	p, err := h.parking.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	return response.Created(c, p)
}

func (h *ParkingHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateParkingSlotRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.UpdateParkingSlotRequest(&req); err != nil {
		return err
	}
	p, err := h.parking.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

func (h *ParkingHandler) Delete(c fiber.Ctx) error {
	if err := h.parking.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "parking slot deleted")
}
