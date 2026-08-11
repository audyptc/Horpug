package delivery

import (
	"time"

	aldomain "apigofiberhorpug/internal/feature/activitylog/domain"
	alusecase "apigofiberhorpug/internal/feature/activitylog/usecase"
	"apigofiberhorpug/internal/feature/electricmeter/domain"
	"apigofiberhorpug/internal/feature/electricmeter/usecase"
	"apigofiberhorpug/internal/shared/http/apierror"
	"apigofiberhorpug/internal/shared/http/httputil"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type ElectricMeterHandler struct {
	uc          *usecase.ElectricMeterUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewElectricMeterHandler(uc *usecase.ElectricMeterUseCase, activityLog *alusecase.ActivityLogUseCase) *ElectricMeterHandler {
	return &ElectricMeterHandler{uc: uc, activityLog: activityLog}
}

// List godoc
// @Summary      รายการค่ามิเตอร์ไฟฟ้า
// @Tags         electric-meters
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.ElectricMeter
// @Router       /electric-meters [get]
func (h *ElectricMeterHandler) List(c fiber.Ctx) error {
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
// @Summary      ดูค่ามิเตอร์ไฟฟ้าตาม ID
// @Tags         electric-meters
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Electric Meter Reading ID"
// @Success      200  {object}  domain.ElectricMeterDetail
// @Router       /electric-meters/{id} [get]
func (h *ElectricMeterHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.uc.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

// GetLatestByRoomID godoc
// @Summary      ดูค่ามิเตอร์ไฟฟ้าล่าสุดของห้อง
// @Tags         electric-meters
// @Security     ApiKeyAuth
// @Produce      json
// @Param        room_id query string true "Room ID"
// @Param        billing_month query string false "Billing month (YYYY-MM-DD)"
// @Success      200  {object}  domain.ElectricMeterDetail
// @Router       /electric-meters/latest [get]
func (h *ElectricMeterHandler) GetLatestByRoomID(c fiber.Ctx) error {
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
// @Summary      บันทึกค่ามิเตอร์ไฟฟ้า
// @Tags         electric-meters
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateElectricMeterRequest true "Electric meter payload"
// @Success      201  {object}  domain.ElectricMeter
// @Router       /electric-meters [post]
func (h *ElectricMeterHandler) Create(c fiber.Ctx) error {
	var req domain.CreateElectricMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateElectricMeterRequest(&req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.uc.Create(c.Context(), dormitoryID, &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityCreate, "electric_meter", d.ID, d)
	return response.Created(c, d)
}

// Update godoc
// @Summary      แก้ไขค่ามิเตอร์ไฟฟ้า
// @Tags         electric-meters
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Electric Meter Reading ID"
// @Param        body body domain.UpdateElectricMeterRequest true "Electric meter payload"
// @Success      200  {object}  domain.ElectricMeter
// @Router       /electric-meters/{id} [put]
func (h *ElectricMeterHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateElectricMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.uc.Update(c.Context(), dormitoryID, c.Params("id"), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityUpdate, "electric_meter", d.ID, d)
	return response.OK(c, d)
}

// Delete godoc
// @Summary      ลบค่ามิเตอร์ไฟฟ้า
// @Tags         electric-meters
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Electric Meter Reading ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /electric-meters/{id} [delete]
func (h *ElectricMeterHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.uc.Delete(c.Context(), dormitoryID, id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityDelete, "electric_meter", id, nil)
	return response.Message(c, "electric meter reading deleted")
}
