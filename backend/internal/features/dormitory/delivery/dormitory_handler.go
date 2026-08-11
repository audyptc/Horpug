package delivery

import (
	aldomain "apigofiberhorpug/internal/features/activitylog/domain"
	alusecase "apigofiberhorpug/internal/features/activitylog/usecase"
	"apigofiberhorpug/internal/features/dormitory/domain"
	"apigofiberhorpug/internal/features/dormitory/usecase"
	"apigofiberhorpug/internal/shared/http/apierror"
	"apigofiberhorpug/internal/shared/http/httputil"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type DormitoryHandler struct {
	dormitories *usecase.DormitoryUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewDormitoryHandler(dormitories *usecase.DormitoryUseCase, activityLog *alusecase.ActivityLogUseCase) *DormitoryHandler {
	return &DormitoryHandler{dormitories: dormitories, activityLog: activityLog}
}

// List godoc
// @Summary      รายชื่อหอพัก (สาขา)
// @Tags         dormitories
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Dormitory
// @Router       /dormitories [get]
func (h *DormitoryHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitories, total, err := h.dormitories.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, dormitories, page, perPage, total)
}

// Mine godoc
// @Summary      รายชื่อหอพักที่ผู้ใช้ปัจจุบันเข้าถึงได้
// @Tags         dormitories
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Dormitory
// @Router       /dormitories/mine [get]
func (h *DormitoryHandler) Mine(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	roleName, _ := c.Locals("role_name").(string)
	dormitories, err := h.dormitories.ListAccessible(c.Context(), userID, roleName)
	if err != nil {
		return err
	}
	return response.OK(c, dormitories)
}

// GetForUser godoc
// @Summary      ดูรายการหอพัก/บทบาทที่ผู้ใช้ถูกกำหนดสิทธิ์
// @Tags         dormitories
// @Security     ApiKeyAuth
// @Produce      json
// @Param        userId path string true "User ID"
// @Success      200  {array}  domain.DormitoryRoleAssignment
// @Router       /dormitories/users/{userId} [get]
func (h *DormitoryHandler) GetForUser(c fiber.Ctx) error {
	assignments, err := h.dormitories.ListAssignmentsForUser(c.Context(), c.Params("userId"))
	if err != nil {
		return err
	}
	return response.OK(c, assignments)
}

// GetByID godoc
// @Summary      ดูข้อมูลหอพักตาม ID
// @Tags         dormitories
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Dormitory ID"
// @Success      200  {object}  domain.Dormitory
// @Router       /dormitories/{id} [get]
func (h *DormitoryHandler) GetByID(c fiber.Ctx) error {
	dormitory, err := h.dormitories.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, dormitory)
}

// Create godoc
// @Summary      สร้างหอพัก (สาขา)
// @Tags         dormitories
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateDormitoryRequest true "Dormitory payload"
// @Success      201  {object}  domain.Dormitory
// @Router       /dormitories [post]
func (h *DormitoryHandler) Create(c fiber.Ctx) error {
	var req domain.CreateDormitoryRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateDormitoryRequest(&req); err != nil {
		return err
	}
	dormitory, err := h.dormitories.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityCreate, "dormitory", dormitory.ID, dormitory)
	return response.Created(c, dormitory)
}

// Update godoc
// @Summary      แก้ไขหอพัก (สาขา)
// @Tags         dormitories
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Dormitory ID"
// @Param        body body domain.UpdateDormitoryRequest true "Dormitory payload"
// @Success      200  {object}  domain.Dormitory
// @Router       /dormitories/{id} [put]
func (h *DormitoryHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateDormitoryRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	dormitory, err := h.dormitories.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityUpdate, "dormitory", dormitory.ID, dormitory)
	return response.OK(c, dormitory)
}

// Delete godoc
// @Summary      ลบหอพัก (สาขา)
// @Tags         dormitories
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Dormitory ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /dormitories/{id} [delete]
func (h *DormitoryHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.dormitories.Delete(c.Context(), id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityDelete, "dormitory", id, nil)
	return response.Message(c, "dormitory deleted")
}

// AssignToUser godoc
// @Summary      กำหนดหอพักที่ผู้ใช้เข้าถึงได้ (แทนที่ทั้งหมด)
// @Tags         dormitories
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        userId path string true "User ID"
// @Param        body body domain.AssignDormitoriesRequest true "Dormitory assignment payload"
// @Success      200  {object}  map[string]interface{}
// @Router       /dormitories/users/{userId} [put]
func (h *DormitoryHandler) AssignToUser(c fiber.Ctx) error {
	var req domain.AssignDormitoriesRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	userID := c.Params("userId")
	if err := h.dormitories.AssignDormitoriesToUser(c.Context(), userID, &req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityUpdate, "user_dormitory_roles", userID, req)
	return response.Message(c, "dormitories assigned")
}
