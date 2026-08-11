package delivery

import (
	aldomain "apigofiberhorpug/internal/features/activitylog/domain"
	alusecase "apigofiberhorpug/internal/features/activitylog/usecase"
	"apigofiberhorpug/internal/features/user/domain"
	"apigofiberhorpug/internal/features/user/usecase"
	"apigofiberhorpug/internal/shared/http/apierror"
	"apigofiberhorpug/internal/shared/http/httputil"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	users       *usecase.UserUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewUserHandler(users *usecase.UserUseCase, activityLog *alusecase.ActivityLogUseCase) *UserHandler {
	return &UserHandler{users: users, activityLog: activityLog}
}

// List godoc
// @Summary      รายชื่อผู้ใช้
// @Tags         users
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.User
// @Router       /users [get]
func (h *UserHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	users, total, err := h.users.List(c.Context(), perPage, offset)
	if err != nil {
		return err
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
		return err
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
	if err := validateCreateUserRequest(&req); err != nil {
		return err
	}
	user, err := h.users.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityCreate, "user", user.ID, user)
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
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityUpdate, "user", user.ID, user)
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
	id := c.Params("id")
	if err := h.users.Delete(c.Context(), id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityDelete, "user", id, nil)
	return response.Message(c, "user deleted")
}

// AssignRole godoc
// @Summary      กำหนดบทบาทให้ผู้ใช้
// @Tags         users
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        body body domain.AssignRoleRequest true "Role assignment payload"
// @Success      200  {object}  map[string]interface{}
// @Router       /users/{id}/role [put]
func (h *UserHandler) AssignRole(c fiber.Ctx) error {
	var req domain.AssignRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateAssignRoleRequest(&req); err != nil {
		return err
	}
	id := c.Params("id")
	if err := h.users.AssignRole(c.Context(), id, &req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityUpdate, "user_role", id, &req)
	return response.Message(c, "role assigned successfully")
}
