package http

import (
	"context"
	"errors"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"
	paymentdomain "apihorpug/internal/features/payment/domain"
	portaldomain "apihorpug/internal/features/tenantportal/domain"
	portalusecase "apihorpug/internal/features/tenantportal/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *portalusecase.Service
}

func NewHandler(usecase *portalusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

type sessionRequest struct {
	IDToken string `json:"id_token"`
}

type repairRequest struct {
	RoomID      uuid.UUID `json:"room_id"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
}

func portalError(err error, fallback string) error {
	switch {
	case errors.Is(err, portaldomain.ErrInvalidLineToken):
		return apierror.Unauthorized(err.Error()).WithSlug("invalid_line_token")
	case errors.Is(err, portaldomain.ErrNotLinked):
		return apierror.Forbidden(err.Error()).WithSlug("tenant_not_linked")
	case errors.Is(err, portaldomain.ErrRoomNotRented), errors.Is(err, portaldomain.ErrInvalidRepair):
		return apierror.BadRequest(err.Error())
	case errors.Is(err, portaldomain.ErrRepairNotFound), errors.Is(err, invoicedomain.ErrInvoiceNotFound),
		errors.Is(err, paymentdomain.ErrPaymentNotFound):
		return apierror.NotFound("not found")
	case errors.Is(err, portaldomain.ErrRepairNotPending):
		return apierror.Conflict(err.Error()).WithSlug("repair_not_pending")
	}
	return apierror.Internal(fallback)
}

func withTimeout(c fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 10*time.Second)
}

// StartSession godoc
// @Summary Sign a tenant in from the LIFF page
// @Description Exchanges a LIFF id token for a short-lived tenant token. The LINE account must be linked to an active tenant.
// @Tags tenant-portal
// @Accept json
// @Produce json
// @Param request body sessionRequest true "LIFF id token"
// @Success 200 {object} portaldomain.Session
// @Failure 401 {object} apierror.Error
// @Failure 403 {object} apierror.Error
// @Router /public/tenant-portal/session [post]
func (h *Handler) StartSession(c fiber.Ctx) error {
	var req sessionRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	ctx, cancel := withTimeout(c)
	defer cancel()

	session, err := h.usecase.StartSession(ctx, req.IDToken)
	if err != nil {
		return portalError(err, "failed to sign in")
	}
	return apiresponse.OK(c, session)
}

// Profile godoc
// @Summary The signed-in tenant and the rooms they rent
// @Tags tenant-portal
// @Produce json
// @Success 200 {object} portaldomain.Profile
// @Security TenantAuth
// @Router /tenant/me [get]
func (h *Handler) Profile(c fiber.Ctx) error {
	tenantID, _ := middleware.TenantID(c)
	ctx, cancel := withTimeout(c)
	defer cancel()
	profile, err := h.usecase.Profile(ctx, tenantID)
	if err != nil {
		return portalError(err, "failed to load profile")
	}
	return apiresponse.OK(c, profile)
}

// Invoices godoc
// @Summary The signed-in tenant's invoices
// @Tags tenant-portal
// @Produce json
// @Success 200 {array} portaldomain.Invoice
// @Security TenantAuth
// @Router /tenant/invoices [get]
func (h *Handler) Invoices(c fiber.Ctx) error {
	tenantID, _ := middleware.TenantID(c)
	ctx, cancel := withTimeout(c)
	defer cancel()
	invoices, err := h.usecase.Invoices(ctx, tenantID)
	if err != nil {
		return portalError(err, "failed to load invoices")
	}
	return apiresponse.OK(c, invoices)
}

// InvoiceDocument godoc
// @Summary One of the tenant's invoices with its payments and PromptPay QR
// @Tags tenant-portal
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} invoicedomain.Document
// @Failure 404 {object} apierror.Error
// @Security TenantAuth
// @Router /tenant/invoices/{id} [get]
func (h *Handler) InvoiceDocument(c fiber.Ctx) error {
	tenantID, _ := middleware.TenantID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid invoice id")
	}
	ctx, cancel := withTimeout(c)
	defer cancel()
	doc, err := h.usecase.InvoiceDocument(ctx, tenantID, id)
	if err != nil {
		return portalError(err, "failed to load invoice")
	}
	return apiresponse.OK(c, doc)
}

// Receipt godoc
// @Summary A receipt for one of the tenant's own payments
// @Tags tenant-portal
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} paymentdomain.Receipt
// @Failure 404 {object} apierror.Error
// @Security TenantAuth
// @Router /tenant/payments/{id}/receipt [get]
func (h *Handler) Receipt(c fiber.Ctx) error {
	tenantID, _ := middleware.TenantID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid payment id")
	}
	ctx, cancel := withTimeout(c)
	defer cancel()
	receipt, err := h.usecase.Receipt(ctx, tenantID, id)
	if err != nil {
		return portalError(err, "failed to load receipt")
	}
	return apiresponse.OK(c, receipt)
}

// RepairRequests godoc
// @Summary Repairs the signed-in tenant reported
// @Tags tenant-portal
// @Produce json
// @Success 200 {array} portaldomain.RepairRequest
// @Security TenantAuth
// @Router /tenant/repair-requests [get]
func (h *Handler) RepairRequests(c fiber.Ctx) error {
	tenantID, _ := middleware.TenantID(c)
	ctx, cancel := withTimeout(c)
	defer cancel()
	list, err := h.usecase.RepairRequests(ctx, tenantID)
	if err != nil {
		return portalError(err, "failed to load repair requests")
	}
	return apiresponse.OK(c, list)
}

// CreateRepairRequest godoc
// @Summary Report a repair for a room the tenant rents
// @Tags tenant-portal
// @Accept json
// @Produce json
// @Param request body repairRequest true "Room, category and description"
// @Success 201 {object} portaldomain.RepairRequest
// @Failure 400 {object} apierror.Error
// @Security TenantAuth
// @Router /tenant/repair-requests [post]
func (h *Handler) CreateRepairRequest(c fiber.Ctx) error {
	tenantID, _ := middleware.TenantID(c)
	var req repairRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	ctx, cancel := withTimeout(c)
	defer cancel()
	created, err := h.usecase.CreateRepairRequest(ctx, tenantID, req.RoomID, req.Category, req.Description)
	if err != nil {
		return portalError(err, "failed to report repair")
	}
	return apiresponse.Created(c, created)
}

// CancelRepairRequest godoc
// @Summary Withdraw a pending repair request
// @Tags tenant-portal
// @Produce json
// @Param id path string true "Repair request ID"
// @Success 200 {object} portaldomain.RepairRequest
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Security TenantAuth
// @Router /tenant/repair-requests/{id}/cancel [post]
func (h *Handler) CancelRepairRequest(c fiber.Ctx) error {
	tenantID, _ := middleware.TenantID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid repair request id")
	}
	ctx, cancel := withTimeout(c)
	defer cancel()
	cancelled, err := h.usecase.CancelRepairRequest(ctx, tenantID, id)
	if err != nil {
		return portalError(err, "failed to cancel repair request")
	}
	return apiresponse.OK(c, cancelled)
}

// Announcements godoc
// @Summary Announcements of the dormitories the tenant lives in
// @Tags tenant-portal
// @Produce json
// @Success 200 {array} portaldomain.Announcement
// @Security TenantAuth
// @Router /tenant/announcements [get]
func (h *Handler) Announcements(c fiber.Ctx) error {
	tenantID, _ := middleware.TenantID(c)
	ctx, cancel := withTimeout(c)
	defer cancel()
	list, err := h.usecase.Announcements(ctx, tenantID)
	if err != nil {
		return portalError(err, "failed to load announcements")
	}
	return apiresponse.OK(c, list)
}
