package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	aldomain "apigofiberhorpug/internal/feature/activitylog/domain"
	alusecase "apigofiberhorpug/internal/feature/activitylog/usecase"
	"apigofiberhorpug/internal/feature/dormitory/domain"
	"apigofiberhorpug/internal/feature/dormitory/usecase"

	"github.com/gofiber/fiber/v3"
)

type DormitoryHandler struct {
	dormitories *usecase.DormitoryUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewDormitoryHandler(dormitories *usecase.DormitoryUseCase, activityLog *alusecase.ActivityLogUseCase) *DormitoryHandler {
	return &DormitoryHandler{dormitories: dormitories, activityLog: activityLog}
}

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

// Mine returns the dormitories the current user can access, used by the frontend branch switcher.
func (h *DormitoryHandler) Mine(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	roleName, _ := c.Locals("role_name").(string)
	dormitories, err := h.dormitories.ListAccessible(c.Context(), userID, roleName)
	if err != nil {
		return err
	}
	return response.OK(c, dormitories)
}

func (h *DormitoryHandler) GetByID(c fiber.Ctx) error {
	dormitory, err := h.dormitories.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, dormitory)
}

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

func (h *DormitoryHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.dormitories.Delete(c.Context(), id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityDelete, "dormitory", id, nil)
	return response.Message(c, "dormitory deleted")
}

// AssignToUser replaces the full set of dormitories a given user can access.
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
	h.activityLog.Log(c.Context(), actorID, aldomain.ActivityUpdate, "user_dormitories", userID, req)
	return response.Message(c, "dormitories assigned")
}
