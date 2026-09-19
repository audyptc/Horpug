package http

import (
	"context"
	"strings"
	"time"

	dashboardusecase "apihorpug/internal/features/dashboard/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *dashboardusecase.Service
}

func NewHandler(usecase *dashboardusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

// GetSummary godoc
// @Summary Get dashboard summary
// @Description Returns headline counts over the dormitories the caller manages (every dormitory for roles with full access), or over a single one when dormitory_id is given. A dormitory the caller has no access to counts as zero rather than being disclosed. A section is null when the caller's role cannot read the menu it comes from.
// @Tags dashboard
// @Produce json
// @Param dormitory_id query string false "Narrow the counts to one dormitory"
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

	var dormitoryID *uuid.UUID
	if raw := strings.TrimSpace(c.Query("dormitory_id")); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return apierror.BadRequest("invalid dormitory_id")
		}
		dormitoryID = &parsed
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	summary, err := h.usecase.GetSummary(ctx, requesterID, dormitoryID)
	if err != nil {
		return apierror.Internal("failed to get dashboard summary")
	}

	return apiresponse.OK(c, summary)
}
