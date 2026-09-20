package http

import (
	"context"
	"errors"
	"mime"
	"time"

	documentdomain "apihorpug/internal/features/document/domain"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// File godoc
// @Summary Download a document's stored file
// @Description Returns the bytes of a file uploaded for the document. Requires access to the document's dormitory; documents that only link to an external file have nothing stored and return 404.
// @Tags documents
// @Produce octet-stream
// @Param id path string true "Document ID"
// @Success 200 {file} binary
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /documents/{id}/file [get]
func (h *Handler) File(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid document id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	document, data, err := h.usecase.GetFile(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, documentdomain.ErrDocumentNotFound) || errors.Is(err, documentdomain.ErrFileNotFound) {
			return apierror.NotFound("document file not found")
		}
		return apierror.Internal("failed to read document file")
	}

	// Always an attachment, and never sniffed: the file is user-supplied
	// content coming from our own origin, so it must not be rendered by the
	// browser as a page. The UI previews it from a blob it fetches itself.
	c.Set(fiber.HeaderContentType, document.FileMime)
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	if disposition := mime.FormatMediaType("attachment", map[string]string{"filename": document.FileName}); disposition != "" {
		c.Set(fiber.HeaderContentDisposition, disposition)
	} else {
		c.Set(fiber.HeaderContentDisposition, "attachment")
	}

	return c.Send(data)
}
