package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"
	"apigofiberhorpug/internal/validator"

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
	list, total, err := h.announcement.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *AnnouncementHandler) GetByID(c fiber.Ctx) error {
	a, err := h.announcement.GetByID(c.Context(), c.Params("id"))
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
	if err := validator.CreateAnnouncementRequest(&req); err != nil {
		return err
	}
	a, err := h.announcement.Create(c.Context(), &req)
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
	if err := validator.UpdateAnnouncementRequest(&req); err != nil {
		return err
	}
	a, err := h.announcement.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, a)
}

func (h *AnnouncementHandler) Delete(c fiber.Ctx) error {
	if err := h.announcement.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "announcement deleted")
}
