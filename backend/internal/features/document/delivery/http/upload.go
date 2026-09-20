package http

import (
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"time"

	documentdomain "apihorpug/internal/features/document/domain"
	documentusecase "apihorpug/internal/features/document/usecase"
	"apihorpug/internal/http/apierror"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// uploadError maps the file-validation failures to client errors, each with a
// slug so the UI can show a localized message. It returns nil for any other
// error.
func uploadError(err error) *apierror.Error {
	switch {
	case errors.Is(err, documentdomain.ErrFileTooLarge):
		return apierror.New(fiber.StatusRequestEntityTooLarge, "file is too large").WithSlug("file_too_large")
	case errors.Is(err, documentdomain.ErrUnsupportedFileType):
		return apierror.BadRequest("unsupported file type").WithSlug("unsupported_file_type")
	case errors.Is(err, documentdomain.ErrEmptyFile):
		return apierror.BadRequest("file is empty").WithSlug("empty_file")
	}
	return nil
}

// readUpload reads the optional "file" part of a multipart form. A part
// larger than the limit is rejected without reading the rest of it.
func readUpload(form *multipart.Form) (*documentusecase.FileUpload, error) {
	headers := form.File["file"]
	if len(headers) == 0 {
		return nil, nil
	}

	header := headers[0]
	if header.Size > documentusecase.MaxUploadSize {
		return nil, uploadError(documentdomain.ErrFileTooLarge)
	}

	file, err := header.Open()
	if err != nil {
		return nil, apierror.BadRequest("invalid file upload")
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, documentusecase.MaxUploadSize+1))
	if err != nil {
		return nil, apierror.BadRequest("invalid file upload")
	}
	return &documentusecase.FileUpload{Name: header.Filename, Data: data}, nil
}

func formValue(form *multipart.Form, key string) (string, bool) {
	values := form.Value[key]
	if len(values) == 0 {
		return "", false
	}
	return strings.TrimSpace(values[0]), true
}

func formUUID(form *multipart.Form, key string) (*uuid.UUID, error) {
	raw, ok := formValue(form, key)
	if !ok || raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, apierror.BadRequest("invalid " + key)
	}
	return &id, nil
}

func formTime(form *multipart.Form, key string) (*time.Time, error) {
	raw, ok := formValue(form, key)
	if !ok || raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, apierror.BadRequest("invalid " + key)
	}
	return &t, nil
}

// parseCreateRequest accepts either a JSON body (a document that links to an
// externally hosted file) or a multipart form with an uploaded "file" part.
func parseCreateRequest(c fiber.Ctx) (createDocumentRequest, *documentusecase.FileUpload, error) {
	var req createDocumentRequest

	if !c.IsMultipart() {
		if err := c.Bind().Body(&req); err != nil {
			return req, nil, apierror.BadRequest("invalid request body")
		}
		return req, nil, nil
	}

	form, err := c.MultipartForm()
	if err != nil {
		return req, nil, apierror.BadRequest("invalid request body")
	}

	dormitoryID, err := formUUID(form, "dormitory_id")
	if err != nil {
		return req, nil, err
	}
	if dormitoryID != nil {
		req.DormitoryID = *dormitoryID
	}
	if req.TenantID, err = formUUID(form, "tenant_id"); err != nil {
		return req, nil, err
	}
	if req.RoomID, err = formUUID(form, "room_id"); err != nil {
		return req, nil, err
	}
	if req.UploadedDate, err = formTime(form, "uploaded_date"); err != nil {
		return req, nil, err
	}

	req.Name, _ = formValue(form, "name")
	category, _ := formValue(form, "category")
	req.Category = documentdomain.DocumentCategory(category)
	req.FileURL, _ = formValue(form, "file_url")
	req.Note, _ = formValue(form, "note")

	upload, err := readUpload(form)
	if err != nil {
		return req, nil, err
	}
	return req, upload, nil
}

// parseUpdateRequest is the update counterpart of parseCreateRequest. As with
// the JSON body, a field that is absent (or an empty id) is left unchanged.
func parseUpdateRequest(c fiber.Ctx) (updateDocumentRequest, *documentusecase.FileUpload, error) {
	var req updateDocumentRequest

	if !c.IsMultipart() {
		if err := c.Bind().Body(&req); err != nil {
			return req, nil, apierror.BadRequest("invalid request body")
		}
		return req, nil, nil
	}

	form, err := c.MultipartForm()
	if err != nil {
		return req, nil, apierror.BadRequest("invalid request body")
	}

	if req.TenantID, err = formUUID(form, "tenant_id"); err != nil {
		return req, nil, err
	}
	if req.RoomID, err = formUUID(form, "room_id"); err != nil {
		return req, nil, err
	}
	if req.UploadedDate, err = formTime(form, "uploaded_date"); err != nil {
		return req, nil, err
	}

	if value, ok := formValue(form, "name"); ok {
		req.Name = &value
	}
	if value, ok := formValue(form, "category"); ok {
		category := documentdomain.DocumentCategory(value)
		req.Category = &category
	}
	if value, ok := formValue(form, "file_url"); ok {
		req.FileURL = &value
	}
	if value, ok := formValue(form, "note"); ok {
		req.Note = &value
	}

	upload, err := readUpload(form)
	if err != nil {
		return req, nil, err
	}
	return req, upload, nil
}
