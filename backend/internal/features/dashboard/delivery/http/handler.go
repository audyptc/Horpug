package http

import (
	"context"
	"time"

	dashboardusecase "apihorpug/internal/features/dashboard/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	usecase *dashboardusecase.Service
}

func NewHandler(usecase *dashboardusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

// GetSummary godoc
// @Summary Get dashboard summary
// @Description Returns headline counts over the dormitories the caller manages (every dormitory for roles with full access). A section is null when the caller's role cannot read the menu it comes from.
// @Tags dashboard
// @Produce json
// @Success 200 {object} dashboarddomain.Summary
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /dashboard/summary [get]
func (h *Handler) GetSummary(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	summary, err := h.usecase.GetSummary(ctx, requesterID)
	if err != nil {
		return apierror.Internal("failed to get dashboard summary")
	}

	return apiresponse.OK(c, summary)
}
