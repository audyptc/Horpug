package delivery

import (
	"apigofiberhorpug/internal/features/announcement/domain"
	"apigofiberhorpug/internal/features/announcement/usecase"
	"apigofiberhorpug/internal/shared/http/apierror"
	"apigofiberhorpug/internal/shared/http/httputil"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type AnnouncementHandler struct {
	announcement *usecase.AnnouncementUseCase
}

func NewAnnouncementHandler(announcement *usecase.AnnouncementUseCase) *AnnouncementHandler {
	return &AnnouncementHandler{announcement: announcement}
}

// List godoc
// @Summary      รายการประกาศ
// @Tags         announcements
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Announcement
// @Router       /announcements [get]
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

// GetByID godoc
// @Summary      ดูข้อมูลประกาศตาม ID
// @Tags         announcements
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Announcement ID"
// @Success      200  {object}  domain.Announcement
// @Router       /announcements/{id} [get]
func (h *AnnouncementHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	a, err := h.announcement.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, a)
}

// Create godoc
// @Summary      สร้างประกาศ
// @Tags         announcements
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateAnnouncementRequest true "Announcement payload"
// @Success      201  {object}  domain.Announcement
// @Router       /announcements [post]
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

// Update godoc
// @Summary      แก้ไขประกาศ
// @Tags         announcements
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Announcement ID"
// @Param        body body domain.UpdateAnnouncementRequest true "Announcement payload"
// @Success      200  {object}  domain.Announcement
// @Router       /announcements/{id} [put]
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

// Delete godoc
// @Summary      ลบประกาศ
// @Tags         announcements
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Announcement ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /announcements/{id} [delete]
func (h *AnnouncementHandler) Delete(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.announcement.Delete(c.Context(), dormitoryID, c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "announcement deleted")
}
