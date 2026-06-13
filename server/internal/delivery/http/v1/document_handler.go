package v1

import (
	"os"
	"path/filepath"
	"strings"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"
	"apigofiberhorpug/internal/validator"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type DocumentHandler struct {
	documents     *usecase.DocumentUseCase
	uploadDir     string
	uploadBaseURL string
}

func NewDocumentHandler(documents *usecase.DocumentUseCase, uploadDir, uploadBaseURL string) *DocumentHandler {
	return &DocumentHandler{documents: documents, uploadDir: uploadDir, uploadBaseURL: uploadBaseURL}
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
	old, err := h.documents.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	d, err := h.documents.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	if old.FileURL != "" && old.FileURL != req.FileURL {
		h.removeUploadedFile(old.FileURL)
	}
	return response.OK(c, d)
}

func (h *DocumentHandler) removeUploadedFile(fileURL string) {
	prefix := h.uploadBaseURL + "/uploads/"
	if strings.HasPrefix(fileURL, prefix) {
		relPath := strings.TrimPrefix(fileURL, prefix)
		os.Remove(filepath.Join(h.uploadDir, relPath))
	}
}

func (h *DocumentHandler) Delete(c fiber.Ctx) error {
	d, err := h.documents.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	if err := h.documents.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	if d.FileURL != "" {
		h.removeUploadedFile(d.FileURL)
	}
	return response.Message(c, "document deleted")
}

var allowedExtensions = map[string]bool{
	".pdf": true, ".jpg": true, ".jpeg": true,
	".png": true, ".gif": true, ".webp": true,
	".doc": true, ".docx": true,
}

func (h *DocumentHandler) Upload(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return apierror.BadRequest("file is required")
	}
	if file.Size > 10*1024*1024 {
		return apierror.BadRequest("file must not exceed 10 MB")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		return apierror.BadRequest("only PDF, image (JPG/PNG/GIF/WEBP), and Word files are allowed")
	}

	dir := filepath.Join(h.uploadDir, "documents")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return apierror.Internal(err)
	}

	filename := uuid.New().String() + ext
	if err := c.SaveFile(file, filepath.Join(dir, filename)); err != nil {
		return apierror.Internal(err)
	}

	url := h.uploadBaseURL + "/uploads/documents/" + filename
	return response.OK(c, fiber.Map{"url": url})
}
