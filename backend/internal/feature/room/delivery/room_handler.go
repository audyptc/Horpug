package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	aldomain "apigofiberhorpug/internal/feature/activitylog/domain"
	alusecase "apigofiberhorpug/internal/feature/activitylog/usecase"
	"apigofiberhorpug/internal/feature/room/domain"
	"apigofiberhorpug/internal/feature/room/usecase"

	"github.com/gofiber/fiber/v3"
)

type RoomHandler struct {
	rooms       *usecase.RoomUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewRoomHandler(rooms *usecase.RoomUseCase, activityLog *alusecase.ActivityLogUseCase) *RoomHandler {
	return &RoomHandler{rooms: rooms, activityLog: activityLog}
}

func (h *RoomHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	rooms, total, err := h.rooms.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, rooms, page, perPage, total)
}

func (h *RoomHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	room, err := h.rooms.GetByID(c.Context(), dormitoryID, c.Params("id"))
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
	if err := validateCreateRoomRequest(&req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	room, err := h.rooms.Create(c.Context(), dormitoryID, &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityCreate, "room", room.ID, room)
	return response.Created(c, room)
}

func (h *RoomHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateRoomRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	room, err := h.rooms.Update(c.Context(), dormitoryID, c.Params("id"), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityUpdate, "room", room.ID, room)
	return response.OK(c, room)
}

func (h *RoomHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.rooms.Delete(c.Context(), dormitoryID, id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityDelete, "room", id, nil)
	return response.Message(c, "room deleted")
}
