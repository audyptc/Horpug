package delivery

import (
	"time"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	aldomain "apigofiberhorpug/internal/feature/activitylog/domain"
	alusecase "apigofiberhorpug/internal/feature/activitylog/usecase"
	"apigofiberhorpug/internal/feature/watermeter/domain"
	"apigofiberhorpug/internal/feature/watermeter/usecase"

	"github.com/gofiber/fiber/v3"
)

type WaterMeterHandler struct {
	uc          *usecase.WaterMeterUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewWaterMeterHandler(uc *usecase.WaterMeterUseCase, activityLog *alusecase.ActivityLogUseCase) *WaterMeterHandler {
	return &WaterMeterHandler{uc: uc, activityLog: activityLog}
}

// List godoc
// @Summary      รายการค่ามิเตอร์น้ำ
// @Tags         water-meters
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.WaterMeter
// @Router       /water-meters [get]
func (h *WaterMeterHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	list, total, err := h.uc.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

// GetByID godoc
// @Summary      ดูค่ามิเตอร์น้ำตาม ID
// @Tags         water-meters
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Water Meter Reading ID"
// @Success      200  {object}  domain.WaterMeterDetail
// @Router       /water-meters/{id} [get]
func (h *WaterMeterHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.uc.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

// GetLatestByRoomID godoc
// @Summary      ดูค่ามิเตอร์น้ำล่าสุดของห้อง
// @Tags         water-meters
// @Security     ApiKeyAuth
// @Produce      json
// @Param        room_id query string true "Room ID"
// @Param        billing_month query string false "Billing month (YYYY-MM-DD)"
// @Success      200  {object}  domain.WaterMeterDetail
// @Router       /water-meters/latest [get]
func (h *WaterMeterHandler) GetLatestByRoomID(c fiber.Ctx) error {
	roomID := c.Query("room_id")
	if roomID == "" {
		return apierror.BadRequest("room_id is required")
	}
	var billingMonth *time.Time
	if bm := c.Query("billing_month"); bm != "" {
		t, err := time.Parse("2006-01-02", bm)
		if err != nil {
			return apierror.BadRequest("invalid billing_month, expected YYYY-MM-DD")
		}
		billingMonth = &t
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.uc.GetLatestByRoomID(c.Context(), dormitoryID, roomID, billingMonth)
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

// Create godoc
// @Summary      บันทึกค่ามิเตอร์น้ำ
// @Tags         water-meters
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateWaterMeterRequest true "Water meter payload"
// @Success      201  {object}  domain.WaterMeter
// @Router       /water-meters [post]
func (h *WaterMeterHandler) Create(c fiber.Ctx) error {
	var req domain.CreateWaterMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateWaterMeterRequest(&req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.uc.Create(c.Context(), dormitoryID, &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityCreate, "water_meter", d.ID, d)
	return response.Created(c, d)
}

// Update godoc
// @Summary      แก้ไขค่ามิเตอร์น้ำ
// @Tags         water-meters
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Water Meter Reading ID"
// @Param        body body domain.UpdateWaterMeterRequest true "Water meter payload"
// @Success      200  {object}  domain.WaterMeter
// @Router       /water-meters/{id} [put]
func (h *WaterMeterHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateWaterMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.uc.Update(c.Context(), dormitoryID, c.Params("id"), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityUpdate, "water_meter", d.ID, d)
	return response.OK(c, d)
}

// Delete godoc
// @Summary      ลบค่ามิเตอร์น้ำ
// @Tags         water-meters
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Water Meter Reading ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /water-meters/{id} [delete]
func (h *WaterMeterHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.uc.Delete(c.Context(), dormitoryID, id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityDelete, "water_meter", id, nil)
	return response.Message(c, "water meter reading deleted")
}
