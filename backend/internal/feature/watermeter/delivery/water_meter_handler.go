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

func (h *WaterMeterHandler) List(c fiber.Ctx) error {
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

func (h *WaterMeterHandler) GetByID(c fiber.Ctx) error {
	d, err := h.uc.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

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
	d, err := h.uc.GetLatestByRoomID(c.Context(), roomID, billingMonth)
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *WaterMeterHandler) Create(c fiber.Ctx) error {
	var req domain.CreateWaterMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateWaterMeterRequest(&req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	d, err := h.uc.Create(c.Context(), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityCreate, "water_meter", d.ID, d)
	return response.Created(c, d)
}

func (h *WaterMeterHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateWaterMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	actorID, _ := c.Locals("user_id").(string)
	d, err := h.uc.Update(c.Context(), c.Params("id"), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityUpdate, "water_meter", d.ID, d)
	return response.OK(c, d)
}

func (h *WaterMeterHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.uc.Delete(c.Context(), id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityDelete, "water_meter", id, nil)
	return response.Message(c, "water meter reading deleted")
}
