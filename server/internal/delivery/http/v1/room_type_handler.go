package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type RoomTypeHandler struct {
	uc *usecase.RoomTypeUseCase
}

func NewRoomTypeHandler(uc *usecase.RoomTypeUseCase) *RoomTypeHandler {
	return &RoomTypeHandler{uc: uc}
}

func (h *RoomTypeHandler) List(c fiber.Ctx) error {
	types, err := h.uc.List(c.Context())
	if err != nil {
		return err
	}
	return response.OK(c, types)
}

func (h *RoomTypeHandler) GetByID(c fiber.Ctx) error {
	rt, err := h.uc.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, rt)
}

func (h *RoomTypeHandler) Create(c fiber.Ctx) error {
	var req domain.CreateRoomTypeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	rt, err := h.uc.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	return response.Created(c, rt)
}

func (h *RoomTypeHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateRoomTypeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	rt, err := h.uc.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, rt)
}

func (h *RoomTypeHandler) Delete(c fiber.Ctx) error {
	if err := h.uc.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "room type deleted")
}
