package delivery

import (
	"strconv"

	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/analytics/usecase"

	"github.com/gofiber/fiber/v3"
)

type AnalyticsHandler struct {
	analytics *usecase.AnalyticsUseCase
}

func NewAnalyticsHandler(analytics *usecase.AnalyticsUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

func (h *AnalyticsHandler) Summary(c fiber.Ctx) error {
	months := 12
	if m := c.Query("months"); m != "" {
		if v, err := strconv.Atoi(m); err == nil {
			months = v
		}
	}
	s, err := h.analytics.Summary(c.Context(), months)
	if err != nil {
		return err
	}
	return response.OK(c, s)
}
