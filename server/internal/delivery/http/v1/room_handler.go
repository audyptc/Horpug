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

type RoomHandler struct {
	rooms *usecase.RoomUseCase
}

func NewRoomHandler(rooms *usecase.RoomUseCase) *RoomHandler {
	return &RoomHandler{rooms: rooms}
}

func (h *RoomHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	rooms, total, err := h.rooms.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, rooms, page, perPage, total)
}

func (h *RoomHandler) GetByID(c fiber.Ctx) error {
	room, err := h.rooms.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, room)
}

func (h *RoomHandler) Create(c fiber.Ctx) error {
	var req domain.CreateRoomRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreateRoomRequest(&req); err != nil {
		return err
	}
	room, err := h.rooms.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	return response.Created(c, room)
}

func (h *RoomHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateRoomRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	room, err := h.rooms.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, room)
}

func (h *RoomHandler) Delete(c fiber.Ctx) error {
	if err := h.rooms.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "room deleted")
}
