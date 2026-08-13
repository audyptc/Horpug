package http

import (
	"context"
	"errors"
	"strings"
	"time"

	paymentdomain "apihorpug/internal/features/payment/domain"
	paymentusecase "apihorpug/internal/features/payment/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *paymentusecase.Service
}

type createPaymentRequest struct {
	InvoiceID     uuid.UUID                   `json:"invoice_id"`
	Amount        float64                     `json:"amount"`
	PaymentMethod paymentdomain.PaymentMethod `json:"payment_method"`
	PaymentDate   time.Time                   `json:"payment_date"`
	ReferenceNo   string                      `json:"reference_no"`
	Note          string                      `json:"note"`
}

func NewHandler(usecase *paymentusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

func parseUUIDQuery(c fiber.Ctx, name string) (*uuid.UUID, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, apierror.BadRequest("invalid " + name)
	}
	return &id, nil
}

func parseDateQuery(c fiber.Ctx, name string) (*time.Time, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	value, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, apierror.BadRequest("invalid " + name)
	}
	return &value, nil
}

func parseListFilters(c fiber.Ctx) (paymentusecase.ListFilters, error) {
	invoiceID, err := parseUUIDQuery(c, "invoice_id")
	if err != nil {
		return paymentusecase.ListFilters{}, err
	}
	contractID, err := parseUUIDQuery(c, "contract_id")
	if err != nil {
		return paymentusecase.ListFilters{}, err
	}
	roomID, err := parseUUIDQuery(c, "room_id")
	if err != nil {
		return paymentusecase.ListFilters{}, err
	}
	dormitoryID, err := parseUUIDQuery(c, "dormitory_id")
	if err != nil {
		return paymentusecase.ListFilters{}, err
	}
	tenantID, err := parseUUIDQuery(c, "tenant_id")
	if err != nil {
		return paymentusecase.ListFilters{}, err
	}
	dateFrom, err := parseDateQuery(c, "date_from")
	if err != nil {
		return paymentusecase.ListFilters{}, err
	}
	dateTo, err := parseDateQuery(c, "date_to")
	if err != nil {
		return paymentusecase.ListFilters{}, err
	}

	return paymentusecase.ListFilters{
		InvoiceID:   invoiceID,
		ContractID:  contractID,
		RoomID:      roomID,
		DormitoryID: dormitoryID,
		TenantID:    tenantID,
		DateFrom:    dateFrom,
		DateTo:      dateTo,
	}, nil
}

// List godoc
// @Summary List payments
// @Description Returns every payment for roles with full dormitory access, otherwise only payments under dormitories the caller manages. Optionally filter by invoice, contract, room, dormitory, tenant or payment date range.
// @Tags payments
// @Produce json
// @Param invoice_id query string false "Filter by invoice ID"
// @Param contract_id query string false "Filter by contract ID"
// @Param room_id query string false "Filter by room ID"
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param tenant_id query string false "Filter by tenant ID"
// @Param date_from query string false "Filter by payment date, inclusive (YYYY-MM-DD)"
// @Param date_to query string false "Filter by payment date, inclusive (YYYY-MM-DD)"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /payments [get]
func (h *Handler) List(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	filters, err := parseListFilters(c)
	if err != nil {
		return err
	}

	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	payments, total, err := h.usecase.List(ctx, requesterID, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list payments")
	}

	return apiresponse.Paginated(c, payments, page, perPage, total)
}

// Get godoc
// @Summary Get a payment by ID
// @Tags payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} paymentdomain.Payment
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /payments/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid payment id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	payment, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, paymentdomain.ErrPaymentNotFound) {
			return apierror.NotFound("payment not found")
		}
		return apierror.Internal("failed to get payment")
	}

	return apiresponse.OK(c, payment)
}

// Create godoc
// @Summary Record a payment against an invoice
// @Description Records a payment for an invoice. Once the invoice's recorded payments reach its total_amount, the invoice is automatically marked paid.
// @Tags payments
// @Accept json
// @Produce json
// @Param request body createPaymentRequest true "Payment payload"
// @Success 201 {object} paymentdomain.Payment
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /payments [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createPaymentRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	payment, err := h.usecase.Create(ctx, paymentusecase.CreateInput{
		InvoiceID:     req.InvoiceID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		PaymentDate:   req.PaymentDate,
		ReferenceNo:   req.ReferenceNo,
		Note:          req.Note,
		CreatedBy:     &requesterID,
	})
	if err != nil {
		if errors.Is(err, paymentdomain.ErrRequiredPaymentData) {
			return apierror.BadRequest("invoice_id, amount and payment_date are required")
		}
		if errors.Is(err, paymentdomain.ErrInvalidAmount) {
			return apierror.BadRequest("amount must be greater than zero")
		}
		if errors.Is(err, paymentdomain.ErrInvalidMethod) {
			return apierror.BadRequest("invalid payment method")
		}
		if errors.Is(err, paymentdomain.ErrInvoiceNotFound) {
			return apierror.NotFound("invoice not found")
		}
		if errors.Is(err, paymentdomain.ErrInvoiceCancelled) {
			return apierror.BadRequest("cannot record a payment against a cancelled invoice")
		}
		return apierror.Internal("failed to create payment")
	}

	return apiresponse.Created(c, payment)
}

// Delete godoc
// @Summary Delete a payment
// @Description Deletes a payment and re-evaluates the invoice's paid status accordingly.
// @Tags payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /payments/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid payment id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID); err != nil {
		if errors.Is(err, paymentdomain.ErrPaymentNotFound) {
			return apierror.NotFound("payment not found")
		}
		return apierror.Internal("failed to delete payment")
	}

	return apiresponse.Message(c, "payment deleted")
}
