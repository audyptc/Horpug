package delivery

import (
	aldomain "apigofiberhorpug/internal/features/activitylog/domain"
	alusecase "apigofiberhorpug/internal/features/activitylog/usecase"
	"apigofiberhorpug/internal/features/tenant/domain"
	"apigofiberhorpug/internal/features/tenant/usecase"
	"apigofiberhorpug/internal/shared/http/apierror"
	"apigofiberhorpug/internal/shared/http/httputil"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type TenantHandler struct {
	tenants     *usecase.TenantUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewTenantHandler(tenants *usecase.TenantUseCase, activityLog *alusecase.ActivityLogUseCase) *TenantHandler {
	return &TenantHandler{tenants: tenants, activityLog: activityLog}
}

// List godoc
// @Summary      รายชื่อผู้เช่า
// @Tags         tenants
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Tenant
// @Router       /tenants [get]
func (h *TenantHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	tenants, total, err := h.tenants.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, tenants, page, perPage, total)
}

// GetByID godoc
// @Summary      ดูข้อมูลผู้เช่าตาม ID
// @Tags         tenants
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Tenant ID"
// @Success      200  {object}  domain.Tenant
// @Router       /tenants/{id} [get]
func (h *TenantHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	t, err := h.tenants.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, t)
}

// Create godoc
// @Summary      สร้างผู้เช่า
// @Tags         tenants
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateTenantRequest true "Tenant payload"
// @Success      201  {object}  domain.Tenant
// @Router       /tenants [post]
func (h *TenantHandler) Create(c fiber.Ctx) error {
	var req domain.CreateTenantRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateTenantRequest(&req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	t, err := h.tenants.Create(c.Context(), dormitoryID, &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityCreate, "tenant", t.ID, t)
	return response.Created(c, t)
}

// Update godoc
// @Summary      แก้ไขผู้เช่า
// @Tags         tenants
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Tenant ID"
// @Param        body body domain.UpdateTenantRequest true "Tenant payload"
// @Success      200  {object}  domain.Tenant
// @Router       /tenants/{id} [put]
func (h *TenantHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateTenantRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	t, err := h.tenants.Update(c.Context(), dormitoryID, c.Params("id"), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityUpdate, "tenant", t.ID, t)
	return response.OK(c, t)
}

// Delete godoc
// @Summary      ลบผู้เช่า
// @Tags         tenants
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Tenant ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /tenants/{id} [delete]
func (h *TenantHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.tenants.Delete(c.Context(), dormitoryID, id); err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityDelete, "tenant", id, nil)
	return response.Message(c, "tenant deleted")
}
