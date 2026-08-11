package delivery

import (
	"apigofiberhorpug/internal/features/roomtype/domain"
	"apigofiberhorpug/internal/features/roomtype/usecase"
	"apigofiberhorpug/internal/shared/http/apierror"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type RoomTypeHandler struct {
	uc *usecase.RoomTypeUseCase
}

func NewRoomTypeHandler(uc *usecase.RoomTypeUseCase) *RoomTypeHandler {
	return &RoomTypeHandler{uc: uc}
}

// List godoc
// @Summary      รายชื่อประเภทห้องพัก
// @Tags         room-types
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.RoomType
// @Router       /room-types [get]
func (h *RoomTypeHandler) List(c fiber.Ctx) error {
	types, err := h.uc.List(c.Context())
	if err != nil {
		return err
	}
	return response.OK(c, types)
}

// GetByID godoc
// @Summary      ดูข้อมูลประเภทห้องพักตาม ID
// @Tags         room-types
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Room Type ID"
// @Success      200  {object}  domain.RoomType
// @Router       /room-types/{id} [get]
func (h *RoomTypeHandler) GetByID(c fiber.Ctx) error {
	rt, err := h.uc.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, rt)
}

// Create godoc
// @Summary      สร้างประเภทห้องพัก
// @Tags         room-types
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateRoomTypeRequest true "Room type payload"
// @Success      201  {object}  domain.RoomType
// @Router       /room-types [post]
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

// Update godoc
// @Summary      แก้ไขประเภทห้องพัก
// @Tags         room-types
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Room Type ID"
// @Param        body body domain.UpdateRoomTypeRequest true "Room type payload"
// @Success      200  {object}  domain.RoomType
// @Router       /room-types/{id} [put]
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

// Delete godoc
// @Summary      ลบประเภทห้องพัก
// @Tags         room-types
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Room Type ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /room-types/{id} [delete]
func (h *RoomTypeHandler) Delete(c fiber.Ctx) error {
	if err := h.uc.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "room type deleted")
}
