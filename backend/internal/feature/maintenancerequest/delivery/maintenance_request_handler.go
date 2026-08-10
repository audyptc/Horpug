package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/maintenancerequest/domain"
	"apigofiberhorpug/internal/feature/maintenancerequest/usecase"

	"github.com/gofiber/fiber/v3"
)

type MaintenanceRequestHandler struct {
	maintenance *usecase.MaintenanceRequestUseCase
}

func NewMaintenanceRequestHandler(maintenance *usecase.MaintenanceRequestUseCase) *MaintenanceRequestHandler {
	return &MaintenanceRequestHandler{maintenance: maintenance}
}

func (h *MaintenanceRequestHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	list, total, err := h.maintenance.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *MaintenanceRequestHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	m, err := h.maintenance.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, m)
}

func (h *MaintenanceRequestHandler) Create(c fiber.Ctx) error {
	var req domain.CreateMaintenanceRequestRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateMaintenanceRequestRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	m, err := h.maintenance.Create(c.Context(), dormitoryID, &req)
	if err != nil {
		return err
	}
	return response.Created(c, m)
}

func (h *MaintenanceRequestHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateMaintenanceRequestRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateUpdateMaintenanceRequestRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	m, err := h.maintenance.Update(c.Context(), dormitoryID, c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, m)
}

func (h *MaintenanceRequestHandler) Delete(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.maintenance.Delete(c.Context(), dormitoryID, c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "maintenance request deleted")
}
