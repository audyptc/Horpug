package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/parking/domain"
	"apigofiberhorpug/internal/feature/parking/usecase"

	"github.com/gofiber/fiber/v3"
)

type ParkingHandler struct {
	parking *usecase.ParkingUseCase
}

func NewParkingHandler(parking *usecase.ParkingUseCase) *ParkingHandler {
	return &ParkingHandler{parking: parking}
}

// List godoc
// @Summary      รายการที่จอดรถ
// @Tags         parking
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.ParkingSlot
// @Router       /parking [get]
func (h *ParkingHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	list, total, err := h.parking.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

// GetByID godoc
// @Summary      ดูข้อมูลที่จอดรถตาม ID
// @Tags         parking
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Parking Slot ID"
// @Success      200  {object}  domain.ParkingSlot
// @Router       /parking/{id} [get]
func (h *ParkingHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.parking.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

// Create godoc
// @Summary      สร้างที่จอดรถ
// @Tags         parking
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateParkingSlotRequest true "Parking slot payload"
// @Success      201  {object}  domain.ParkingSlot
// @Router       /parking [post]
func (h *ParkingHandler) Create(c fiber.Ctx) error {
	var req domain.CreateParkingSlotRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateParkingSlotRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.parking.Create(c.Context(), dormitoryID, &req)
	if err != nil {
		return err
	}
	return response.Created(c, p)
}

// Update godoc
// @Summary      แก้ไขที่จอดรถ
// @Tags         parking
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Parking Slot ID"
// @Param        body body domain.UpdateParkingSlotRequest true "Parking slot payload"
// @Success      200  {object}  domain.ParkingSlot
// @Router       /parking/{id} [put]
func (h *ParkingHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateParkingSlotRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateUpdateParkingSlotRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.parking.Update(c.Context(), dormitoryID, c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

// Delete godoc
// @Summary      ลบที่จอดรถ
// @Tags         parking
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Parking Slot ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /parking/{id} [delete]
func (h *ParkingHandler) Delete(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.parking.Delete(c.Context(), dormitoryID, c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "parking slot deleted")
}
