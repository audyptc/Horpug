package delivery

import (
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/dashboard/usecase"

	"github.com/gofiber/fiber/v3"
)

type DashboardHandler struct {
	dashboard *usecase.DashboardUseCase
}

func NewDashboardHandler(dashboard *usecase.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{dashboard: dashboard}
}

func (h *DashboardHandler) Summary(c fiber.Ctx) error {
	s, err := h.dashboard.Summary(c.Context())
	if err != nil {
		return err
	}
	return response.OK(c, s)
}
