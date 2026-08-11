package delivery

import (
	"os"
	"path/filepath"
	"strings"

	"apigofiberhorpug/internal/feature/document/domain"
	"apigofiberhorpug/internal/feature/document/usecase"
	"apigofiberhorpug/internal/shared/http/apierror"
	"apigofiberhorpug/internal/shared/http/httputil"
	"apigofiberhorpug/internal/shared/http/response"

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

// List godoc
// @Summary      รายการเอกสาร
// @Tags         documents
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Document
// @Router       /documents [get]
func (h *DocumentHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	list, total, err := h.documents.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

// GetByID godoc
// @Summary      ดูข้อมูลเอกสารตาม ID
// @Tags         documents
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Document ID"
// @Success      200  {object}  domain.DocumentDetail
// @Router       /documents/{id} [get]
func (h *DocumentHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.documents.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

// Create godoc
// @Summary      สร้างเอกสาร
// @Tags         documents
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateDocumentRequest true "Document payload"
// @Success      201  {object}  domain.Document
// @Router       /documents [post]
func (h *DocumentHandler) Create(c fiber.Ctx) error {
	var req domain.CreateDocumentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateDocumentRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.documents.Create(c.Context(), dormitoryID, &req)
	if err != nil {
		return err
	}
	return response.Created(c, d)
}

// Update godoc
// @Summary      แก้ไขเอกสาร
// @Tags         documents
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Document ID"
// @Param        body body domain.UpdateDocumentRequest true "Document payload"
// @Success      200  {object}  domain.Document
// @Router       /documents/{id} [put]
func (h *DocumentHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateDocumentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateUpdateDocumentRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	old, err := h.documents.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	d, err := h.documents.Update(c.Context(), dormitoryID, c.Params("id"), &req)
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

// Delete godoc
// @Summary      ลบเอกสาร
// @Tags         documents
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Document ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /documents/{id} [delete]
func (h *DocumentHandler) Delete(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.documents.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	if err := h.documents.Delete(c.Context(), dormitoryID, c.Params("id")); err != nil {
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

// Upload godoc
// @Summary      อัปโหลดไฟล์เอกสาร
// @Tags         documents
// @Security     ApiKeyAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Document file (PDF, image, or Word, max 10MB)"
// @Success      200  {object}  map[string]interface{}
// @Router       /documents/upload [post]
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
