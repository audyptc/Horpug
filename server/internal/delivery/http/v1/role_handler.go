package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type RoleHandler struct {
	roles *usecase.RoleUseCase
}

func NewRoleHandler(roles *usecase.RoleUseCase) *RoleHandler {
	return &RoleHandler{roles: roles}
}

// List godoc
// @Summary      รายชื่อบทบาท
// @Tags         roles
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Role
// @Router       /roles [get]
func (h *RoleHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := parsePaginationQuery(c)
	if err != nil {
		return err
	}
	roles, total, err := h.roles.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, roles, page, perPage, total)
}

// ListActive godoc
// @Summary      รายชื่อบทบาทที่เปิดใช้งาน
// @Tags         roles
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Role
// @Router       /roles/active [get]
func (h *RoleHandler) ListActive(c fiber.Ctx) error {
	roles, err := h.roles.ListActive(c.Context())
	if err != nil {
		return err
	}
	return response.OK(c, roles)
}

// GetByID godoc
// @Summary      ดูข้อมูลบทบาทตาม ID
// @Tags         roles
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Role ID"
// @Success      200  {object}  domain.Role
// @Router       /roles/{id} [get]
func (h *RoleHandler) GetByID(c fiber.Ctx) error {
	role, err := h.roles.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, role)
}

// Create godoc
// @Summary      สร้างบทบาท
// @Tags         roles
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateRoleRequest true "Role payload"
// @Success      201  {object}  domain.Role
// @Router       /roles [post]
func (h *RoleHandler) Create(c fiber.Ctx) error {
	var req domain.CreateRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	role, err := h.roles.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	return response.Created(c, role)
}

// Update godoc
// @Summary      แก้ไขบทบาท
// @Tags         roles
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID"
// @Param        body body domain.UpdateRoleRequest true "Role payload"
// @Success      200  {object}  domain.Role
// @Router       /roles/{id} [put]
func (h *RoleHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	role, err := h.roles.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, role)
}

// Delete godoc
// @Summary      ลบบทบาท
// @Tags         roles
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Role ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /roles/{id} [delete]
func (h *RoleHandler) Delete(c fiber.Ctx) error {
	if err := h.roles.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "role deleted")
}

// AssignPermissions godoc
// @Summary      กำหนดสิทธิ์ให้บทบาท
// @Tags         roles
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID"
// @Param        body body domain.AssignPermissionsRequest true "Menu-permission assignment payload"
// @Success      200  {object}  map[string]interface{}
// @Router       /roles/{id}/permissions [put]
func (h *RoleHandler) AssignPermissions(c fiber.Ctx) error {
	var req domain.AssignPermissionsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := h.roles.AssignPermissions(c.Context(), c.Params("id"), &req); err != nil {
		return err
	}
	return response.Message(c, "permissions assigned successfully")
}
