package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	users *usecase.UserUseCase
}

func NewUserHandler(users *usecase.UserUseCase) *UserHandler {
	return &UserHandler{users: users}
}

// List godoc
// @Summary      รายชื่อผู้ใช้
// @Tags         users
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.User
// @Router       /users [get]
func (h *UserHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := parsePaginationQuery(c)
	if err != nil {
		return err
	}

	users, total, err := h.users.List(c.Context(), perPage, offset)
	if err != nil {
		return apierror.Internal(err)
	}
	return response.Paginated(c, users, page, perPage, total)
}

// GetByID godoc
// @Summary      ดูข้อมูลผู้ใช้ตาม ID
// @Tags         users
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200  {object}  domain.User
// @Router       /users/{id} [get]
func (h *UserHandler) GetByID(c fiber.Ctx) error {
	user, err := h.users.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return apierror.NotFound(err.Error())
	}
	return response.OK(c, user)
}

// Create godoc
// @Summary      สร้างผู้ใช้
// @Tags         users
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateUserRequest true "User payload"
// @Success      201  {object}  domain.User
// @Router       /users [post]
func (h *UserHandler) Create(c fiber.Ctx) error {
	var req domain.CreateUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	user, err := h.users.Create(c.Context(), &req)
	if err != nil {
		return apierror.BadRequest(err.Error())
	}
	return response.Created(c, user)
}

// Update godoc
// @Summary      แก้ไขผู้ใช้
// @Tags         users
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        body body domain.UpdateUserRequest true "User payload"
// @Success      200  {object}  domain.User
// @Router       /users/{id} [put]
func (h *UserHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	user, err := h.users.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return apierror.BadRequest(err.Error())
	}
	return response.OK(c, user)
}

// Delete godoc
// @Summary      ลบผู้ใช้
// @Tags         users
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(c fiber.Ctx) error {
	if err := h.users.Delete(c.Context(), c.Params("id")); err != nil {
		return apierror.NotFound(err.Error())
	}
	return response.Message(c, "user deleted")
}

// AssignRoles godoc
// @Summary      กำหนดบทบาทให้ผู้ใช้
// @Tags         users
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        body body domain.AssignRolesRequest true "Role assignment payload"
// @Success      200  {object}  map[string]interface{}
// @Router       /users/{id}/roles [put]
func (h *UserHandler) AssignRoles(c fiber.Ctx) error {
	var req domain.AssignRolesRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := h.users.AssignRoles(c.Context(), c.Params("id"), &req); err != nil {
		return apierror.BadRequest(err.Error())
	}
	return response.Message(c, "role assigned successfully")
}
