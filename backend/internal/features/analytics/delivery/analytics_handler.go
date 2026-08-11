package delivery

import (
	"strconv"

	"apigofiberhorpug/internal/features/analytics/domain"
	"apigofiberhorpug/internal/features/analytics/usecase"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type AnalyticsHandler struct {
	analytics *usecase.AnalyticsUseCase
}

func NewAnalyticsHandler(analytics *usecase.AnalyticsUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

// Summary godoc
// @Summary      สรุปข้อมูลเชิงวิเคราะห์
// @Tags         analytics
// @Security     ApiKeyAuth
// @Produce      json
// @Param        months query int false "Number of months to include (default 12)"
// @Success      200  {object}  domain.AnalyticsSummary
// @Router       /analytics/summary [get]
func (h *AnalyticsHandler) Summary(c fiber.Ctx) error {
	months := 12
	if m := c.Query("months"); m != "" {
		if v, err := strconv.Atoi(m); err == nil {
			months = v
		}
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	var s *domain.AnalyticsSummary
	s, err := h.analytics.Summary(c.Context(), dormitoryID, months)
	if err != nil {
		return err
	}
	return response.OK(c, s)
}
