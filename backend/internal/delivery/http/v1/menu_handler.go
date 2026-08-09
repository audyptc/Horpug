package v1

import (
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type MenuHandler struct {
	menus *usecase.MenuUseCase
}

func NewMenuHandler(menus *usecase.MenuUseCase) *MenuHandler {
	return &MenuHandler{menus: menus}
}

// List godoc
// @Summary      รายการเมนู
// @Tags         menus
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Menu
// @Router       /menus [get]
func (h *MenuHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	menus, total, err := h.menus.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, menus, page, perPage, total)
}

// GetByID godoc
// @Summary      ดูเมนูตาม ID
// @Tags         menus
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Menu ID"
// @Success      200  {object}  domain.Menu
// @Router       /menus/{id} [get]
func (h *MenuHandler) GetByID(c fiber.Ctx) error {
	menu, err := h.menus.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, menu)
}
