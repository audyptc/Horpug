package v1

import (
	"time"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	aldomain "apigofiberhorpug/internal/feature/activitylog/domain"
	alusecase "apigofiberhorpug/internal/feature/activitylog/usecase"
	"apigofiberhorpug/internal/usecase"
	"apigofiberhorpug/internal/validator"

	"github.com/gofiber/fiber/v3"
)

type ElectricMeterHandler struct {
	uc          *usecase.ElectricMeterUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewElectricMeterHandler(uc *usecase.ElectricMeterUseCase, activityLog *alusecase.ActivityLogUseCase) *ElectricMeterHandler {
	return &ElectricMeterHandler{uc: uc, activityLog: activityLog}
}

func (h *ElectricMeterHandler) List(c fiber.Ctx) error {
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

func (h *ElectricMeterHandler) GetByID(c fiber.Ctx) error {
	d, err := h.uc.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

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
	d, err := h.uc.GetLatestByRoomID(c.Context(), roomID, billingMonth)
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *ElectricMeterHandler) Create(c fiber.Ctx) error {
	var req domain.CreateElectricMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreateElectricMeterRequest(&req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	d, err := h.uc.Create(c.Context(), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityCreate, "electric_meter", d.ID, d)
	return response.Created(c, d)
}

func (h *ElectricMeterHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateElectricMeterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	actorID, _ := c.Locals("user_id").(string)
	d, err := h.uc.Update(c.Context(), c.Params("id"), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityUpdate, "electric_meter", d.ID, d)
	return response.OK(c, d)
}

func (h *ElectricMeterHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.uc.Delete(c.Context(), id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityDelete, "electric_meter", id, nil)
	return response.Message(c, "electric meter reading deleted")
}
