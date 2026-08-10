package delivery

import (
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/dashboard/domain"
	"apigofiberhorpug/internal/feature/dashboard/usecase"

	"github.com/gofiber/fiber/v3"
)

type DashboardHandler struct {
	dashboard *usecase.DashboardUseCase
}

func NewDashboardHandler(dashboard *usecase.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{dashboard: dashboard}
}

// Summary godoc
// @Summary      สรุปข้อมูลแดชบอร์ด
// @Tags         dashboard
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {object}  domain.DashboardSummary
// @Router       /dashboard/summary [get]
func (h *DashboardHandler) Summary(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	var s *domain.DashboardSummary
	s, err := h.dashboard.Summary(c.Context(), dormitoryID)
	if err != nil {
		return err
	}
	return response.OK(c, s)
}
