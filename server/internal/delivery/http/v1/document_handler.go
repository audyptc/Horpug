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

type DocumentHandler struct {
	documents *usecase.DocumentUseCase
}

func NewDocumentHandler(documents *usecase.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{documents: documents}
}

func (h *DocumentHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	list, total, err := h.documents.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *DocumentHandler) GetByID(c fiber.Ctx) error {
	d, err := h.documents.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *DocumentHandler) Create(c fiber.Ctx) error {
	var req domain.CreateDocumentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreateDocumentRequest(&req); err != nil {
		return err
	}
	d, err := h.documents.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	return response.Created(c, d)
}

func (h *DocumentHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateDocumentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.UpdateDocumentRequest(&req); err != nil {
		return err
	}
	d, err := h.documents.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *DocumentHandler) Delete(c fiber.Ctx) error {
	if err := h.documents.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "document deleted")
}
