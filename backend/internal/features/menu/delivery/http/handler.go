package http

import (
	"context"
	"time"

	menudomain "apihorpug/internal/features/menu/domain"
	menuusecase "apihorpug/internal/features/menu/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"

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
// @Success 200 {array} menudomain.Menu
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /menus [get]
func (h *Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var menus []menudomain.Menu
	menus, err := h.usecase.List(ctx)
	if err != nil {
		return apierror.Internal("failed to list menus")
	}

	return apiresponse.OK(c, menus)
}
