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

type TenantHandler struct {
	tenants     *usecase.TenantUseCase
	activityLog *usecase.ActivityLogUseCase
}

func NewTenantHandler(tenants *usecase.TenantUseCase, activityLog *usecase.ActivityLogUseCase) *TenantHandler {
	return &TenantHandler{tenants: tenants, activityLog: activityLog}
}

func (h *TenantHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	tenants, total, err := h.tenants.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, tenants, page, perPage, total)
}

func (h *TenantHandler) GetByID(c fiber.Ctx) error {
	t, err := h.tenants.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, t)
}

func (h *TenantHandler) Create(c fiber.Ctx) error {
	var req domain.CreateTenantRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreateTenantRequest(&req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	t, err := h.tenants.Create(c.Context(), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.Log(c.Context(), actorID, domain.ActivityCreate, "tenant", t.ID, t)
	return response.Created(c, t)
}

func (h *TenantHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateTenantRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	actorID, _ := c.Locals("user_id").(string)
	t, err := h.tenants.Update(c.Context(), c.Params("id"), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.Log(c.Context(), actorID, domain.ActivityUpdate, "tenant", t.ID, t)
	return response.OK(c, t)
}

func (h *TenantHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	actorID, _ := c.Locals("user_id").(string)
	if err := h.tenants.Delete(c.Context(), id); err != nil {
		return err
	}
	h.activityLog.Log(c.Context(), actorID, domain.ActivityDelete, "tenant", id, nil)
	return response.Message(c, "tenant deleted")
}
