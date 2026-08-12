package http

import (
	"context"
	"time"

	menuusecase "apihorpug/internal/features/menu/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	usecase *menuusecase.Service
}

func NewHandler(usecase *menuusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

// List godoc
// @Summary List menus
// @Tags menus
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /menus [get]
func (h *Handler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	menus, total, err := h.usecase.List(ctx, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list menus")
	}

	return apiresponse.Paginated(c, menus, page, perPage, total)
}
