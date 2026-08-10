package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/announcement/domain"
	"apigofiberhorpug/internal/feature/announcement/usecase"

	"github.com/gofiber/fiber/v3"
)

type AnnouncementHandler struct {
	announcement *usecase.AnnouncementUseCase
}

func NewAnnouncementHandler(announcement *usecase.AnnouncementUseCase) *AnnouncementHandler {
	return &AnnouncementHandler{announcement: announcement}
}

func (h *AnnouncementHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	list, total, err := h.announcement.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *AnnouncementHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	a, err := h.announcement.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, a)
}

func (h *AnnouncementHandler) Create(c fiber.Ctx) error {
	var req domain.CreateAnnouncementRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateAnnouncementRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	a, err := h.announcement.Create(c.Context(), dormitoryID, &req)
	if err != nil {
		return err
	}
	return response.Created(c, a)
}

func (h *AnnouncementHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateAnnouncementRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateUpdateAnnouncementRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	a, err := h.announcement.Update(c.Context(), dormitoryID, c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, a)
}

func (h *AnnouncementHandler) Delete(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.announcement.Delete(c.Context(), dormitoryID, c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "announcement deleted")
}
