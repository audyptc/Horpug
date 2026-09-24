package http

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	paymentdomain "apihorpug/internal/features/payment/domain"
	slipdomain "apihorpug/internal/features/paymentslip/domain"
	slipusecase "apihorpug/internal/features/paymentslip/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *slipusecase.Service
}

func NewHandler(usecase *slipusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

type approveRequest struct {
	Amount      float64    `json:"amount"`
	PaymentDate *time.Time `json:"payment_date"`
	ReferenceNo string     `json:"reference_no"`
}

type rejectRequest struct {
	Reason string `json:"reason"`
}

func slipError(err error, fallback string) error {
	switch {
	case errors.Is(err, slipdomain.ErrSlipNotFound):
		return apierror.NotFound("slip not found")
	case errors.Is(err, slipdomain.ErrInvoiceNotPayable):
		return apierror.Conflict(err.Error()).WithSlug("invoice_not_payable")
	case errors.Is(err, slipdomain.ErrNotPending):
		return apierror.Conflict(err.Error()).WithSlug("slip_not_pending")
	case errors.Is(err, slipdomain.ErrInvalidAmount):
		return apierror.BadRequest(err.Error()).WithSlug("invalid_amount")
	case errors.Is(err, slipdomain.ErrInvalidDate):
		return apierror.BadRequest(err.Error()).WithSlug("invalid_transfer_date")
	case errors.Is(err, slipdomain.ErrReasonRequired):
		return apierror.BadRequest(err.Error())
	case errors.Is(err, slipdomain.ErrEmptyFile), errors.Is(err, slipdomain.ErrUnsupportedFile):
		return apierror.BadRequest(err.Error()).WithSlug("unsupported_file_type")
	case errors.Is(err, slipdomain.ErrFileTooLarge):
		return apierror.New(fiber.StatusRequestEntityTooLarge, err.Error()).WithSlug("file_too_large")
	case errors.Is(err, paymentdomain.ErrPaymentExceedsInvoice):
		return apierror.Conflict(err.Error()).WithSlug("payment_exceeds_invoice")
	case errors.Is(err, paymentdomain.ErrInvoiceCancelled):
		return apierror.Conflict(err.Error()).WithSlug("invoice_not_payable")
	}
	return apierror.Internal(fallback)
}

func withTimeout(c fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 20*time.Second)
}

// ---- tenant side

// Submit godoc
// @Summary Send a transfer slip for one of the tenant's invoices
// @Description Multipart form: file (JPG, PNG or WEBP, up to 5 MB), amount, transfer_date (YYYY-MM-DD), note. The slip waits for staff review.
// @Tags tenant-portal
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 201 {object} slipdomain.Slip
// @Failure 400 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Security TenantAuth
// @Router /tenant/invoices/{id}/slips [post]
func (h *Handler) Submit(c fiber.Ctx) error {
	tenantID, _ := middleware.TenantID(c)
	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid invoice id")
	}
	form, err := c.MultipartForm()
	if err != nil {
		return apierror.BadRequest("invalid form")
	}
	value := func(key string) string {
		if v := form.Value[key]; len(v) > 0 {
			return strings.TrimSpace(v[0])
		}
		return ""
	}

	amount, err := strconv.ParseFloat(value("amount"), 64)
	if err != nil {
		return slipError(slipdomain.ErrInvalidAmount, "")
	}
	transferDate, err := time.Parse("2006-01-02", value("transfer_date"))
	if err != nil {
		return slipError(slipdomain.ErrInvalidDate, "")
	}

	files := form.File["file"]
	if len(files) == 0 {
		return slipError(slipdomain.ErrEmptyFile, "")
	}
	if files[0].Size > slipusecase.MaxSlipSize {
		return slipError(slipdomain.ErrFileTooLarge, "")
	}
	file, err := files[0].Open()
	if err != nil {
		return apierror.BadRequest("invalid file upload")
	}
	defer file.Close()
	image, err := io.ReadAll(io.LimitReader(file, slipusecase.MaxSlipSize+1))
	if err != nil {
		return apierror.BadRequest("invalid file upload")
	}

	ctx, cancel := withTimeout(c)
	defer cancel()
	slip, err := h.usecase.Submit(ctx, slipusecase.SubmitInput{
		TenantID:     tenantID,
		InvoiceID:    invoiceID,
		Amount:       amount,
		TransferDate: transferDate,
		Note:         value("note"),
		Image:        image,
	})
	if err != nil {
		return slipError(err, "failed to send slip")
	}
	return apiresponse.Created(c, slip)
}

// TenantList godoc
// @Summary Slips the tenant sent for one invoice
// @Tags tenant-portal
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {array} slipdomain.Slip
// @Security TenantAuth
// @Router /tenant/invoices/{id}/slips [get]
func (h *Handler) TenantList(c fiber.Ctx) error {
	tenantID, _ := middleware.TenantID(c)
	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid invoice id")
	}
	ctx, cancel := withTimeout(c)
	defer cancel()
	slips, err := h.usecase.ListForTenantInvoice(ctx, tenantID, invoiceID)
	if err != nil {
		return slipError(err, "failed to load slips")
	}
	return apiresponse.OK(c, slips)
}

// TenantCancel godoc
// @Summary Withdraw a slip that hasn't been reviewed yet
// @Tags tenant-portal
// @Produce json
// @Param id path string true "Slip ID"
// @Success 200 {object} slipdomain.Slip
// @Failure 409 {object} apierror.Error
// @Security TenantAuth
// @Router /tenant/slips/{id}/cancel [post]
func (h *Handler) TenantCancel(c fiber.Ctx) error {
	tenantID, _ := middleware.TenantID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid slip id")
	}
	ctx, cancel := withTimeout(c)
	defer cancel()
	slip, err := h.usecase.CancelByTenant(ctx, tenantID, id)
	if err != nil {
		return slipError(err, "failed to cancel slip")
	}
	return apiresponse.OK(c, slip)
}

// ---- staff side

// List godoc
// @Summary Transfer slips sent by tenants
// @Description Slips in the dormitories the requester manages with the given status (default pending, oldest first).
// @Tags payment-slips
// @Produce json
// @Param status query string false "pending, approved, rejected or cancelled"
// @Success 200 {array} slipdomain.Slip
// @Security BearerAuth
// @Router /payment-slips [get]
func (h *Handler) List(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}
	ctx, cancel := withTimeout(c)
	defer cancel()
	slips, err := h.usecase.ListForStaff(ctx, requesterID, slipdomain.Status(c.Query("status")))
	if err != nil {
		return slipError(err, "failed to load slips")
	}
	return apiresponse.OK(c, slips)
}

// Image godoc
// @Summary The slip image
// @Tags payment-slips
// @Produce image/jpeg,image/png,image/webp
// @Param id path string true "Slip ID"
// @Success 200 {file} file
// @Failure 404 {object} apierror.Error
// @Security BearerAuth
// @Router /payment-slips/{id}/image [get]
func (h *Handler) Image(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid slip id")
	}
	ctx, cancel := withTimeout(c)
	defer cancel()
	slip, data, err := h.usecase.Image(ctx, id, requesterID)
	if err != nil {
		return slipError(err, "failed to load slip image")
	}
	// Tenant-supplied content from our own origin: never sniffed, never
	// cached; the UI shows it from a blob it fetches itself.
	c.Set(fiber.HeaderContentType, slip.FileMime)
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	c.Set(fiber.HeaderContentDisposition, "attachment")
	return c.Send(data)
}

// Approve godoc
// @Summary Approve a slip and record it as a transfer payment
// @Description Records the payment (with its receipt number) for the slip's invoice and notifies the tenant on LINE. amount and payment_date default to the slip's own.
// @Tags payment-slips
// @Accept json
// @Produce json
// @Param id path string true "Slip ID"
// @Param request body approveRequest false "Corrections"
// @Success 200 {object} slipdomain.Slip
// @Failure 409 {object} apierror.Error
// @Security BearerAuth
// @Router /payment-slips/{id}/approve [post]
func (h *Handler) Approve(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid slip id")
	}
	var req approveRequest
	if len(c.Body()) > 0 {
		if err := c.Bind().Body(&req); err != nil {
			return apierror.BadRequest("invalid request body")
		}
	}
	input := slipusecase.ApproveInput{Amount: req.Amount, ReferenceNo: req.ReferenceNo}
	if req.PaymentDate != nil {
		input.PaymentDate = *req.PaymentDate
	}
	ctx, cancel := withTimeout(c)
	defer cancel()
	slip, err := h.usecase.Approve(ctx, id, requesterID, input, c.IP())
	if err != nil {
		return slipError(err, "failed to approve slip")
	}
	return apiresponse.OK(c, slip)
}

// Reject godoc
// @Summary Reject a slip with a reason
// @Tags payment-slips
// @Accept json
// @Produce json
// @Param id path string true "Slip ID"
// @Param request body rejectRequest true "Reason shown to the tenant"
// @Success 200 {object} slipdomain.Slip
// @Failure 400 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Security BearerAuth
// @Router /payment-slips/{id}/reject [post]
func (h *Handler) Reject(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid slip id")
	}
	var req rejectRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	ctx, cancel := withTimeout(c)
	defer cancel()
	slip, err := h.usecase.Reject(ctx, id, requesterID, req.Reason)
	if err != nil {
		return slipError(err, "failed to reject slip")
	}
	return apiresponse.OK(c, slip)
}
