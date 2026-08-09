package v1

import (
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type PermissionHandler struct {
	perms *usecase.PermissionUseCase
}

func NewPermissionHandler(perms *usecase.PermissionUseCase) *PermissionHandler {
	return &PermissionHandler{perms: perms}
}

// List godoc
// @Summary      รายชื่อสิทธิ์
// @Tags         permissions
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Permission
// @Router       /permissions [get]
func (h *PermissionHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	perms, total, err := h.perms.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, perms, page, perPage, total)
}
