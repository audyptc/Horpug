package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"
	"apigofiberhorpug/internal/validator"

	"github.com/gofiber/fiber/v3"
)

type MaintenanceRequestHandler struct {
	maintenance *usecase.MaintenanceRequestUseCase
}

func NewMaintenanceRequestHandler(maintenance *usecase.MaintenanceRequestUseCase) *MaintenanceRequestHandler {
	return &MaintenanceRequestHandler{maintenance: maintenance}
}

func (h *MaintenanceRequestHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := parsePaginationQuery(c)
	if err != nil {
		return err
	}
	list, total, err := h.maintenance.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *MaintenanceRequestHandler) GetByID(c fiber.Ctx) error {
	m, err := h.maintenance.GetByID(c.Context(), c.Params("id"))
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
	if err := validator.CreateMaintenanceRequestRequest(&req); err != nil {
		return err
	}
	m, err := h.maintenance.Create(c.Context(), &req)
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
	if err := validator.UpdateMaintenanceRequestRequest(&req); err != nil {
		return err
	}
	m, err := h.maintenance.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, m)
}

func (h *MaintenanceRequestHandler) Delete(c fiber.Ctx) error {
	if err := h.maintenance.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "maintenance request deleted")
}
