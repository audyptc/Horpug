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

type createPaymentItemRequest struct {
	PaymentMethod paymentdomain.PaymentMethod `json:"payment_method"`
	Amount        float64                     `json:"amount"`
	ReferenceNo   string                      `json:"reference_no"`
}

type createPaymentRequest struct {
	InvoiceID   uuid.UUID                  `json:"invoice_id"`
	PaymentDate time.Time                  `json:"payment_date"`
	Note        string                     `json:"note"`
	Items       []createPaymentItemRequest `json:"items"`
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

// parseColumnFilters reads the per-column filters, passed as f[<column>]=value.
// An unrecognised column is rejected rather than ignored: silently dropping it
// would return an unfiltered list that looks like a legitimate result.
func parseColumnFilters(c fiber.Ctx) (map[string]string, error) {
	columns := make(map[string]string)

	for key, value := range c.Queries() {
		if !strings.HasPrefix(key, "f[") || !strings.HasSuffix(key, "]") {
			continue
		}

		name := key[len("f[") : len(key)-len("]")]
		if _, ok := paymentusecase.FilterColumns[name]; !ok {
			return nil, apierror.BadRequest("unsupported filter field: " + name)
		}

		if value = strings.TrimSpace(value); value != "" {
			columns[name] = value
		}
	}

	return columns, nil
}

func parseOptionalPaymentMethod(c fiber.Ctx, name string) (*paymentdomain.PaymentMethod, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}

	method := paymentdomain.PaymentMethod(raw)
	if !method.Valid() {
		return nil, apierror.BadRequest(name + " must be a valid payment method")
	}
	return &method, nil
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
	columns, err := parseColumnFilters(c)
	if err != nil {
		return paymentusecase.ListFilters{}, err
	}
	paymentMethod, err := parseOptionalPaymentMethod(c, "payment_method")
	if err != nil {
		return paymentusecase.ListFilters{}, err
	}

	sortKey := paymentusecase.DefaultSortKey
	if raw := strings.TrimSpace(c.Query("sort")); raw != "" {
		if _, ok := paymentusecase.SortColumns[raw]; !ok {
			return paymentusecase.ListFilters{}, apierror.BadRequest("unsupported sort field")
		}
		sortKey = raw
	}

	// Newest-first is the useful default for an unsorted listing, but an
	// explicit sort field reads more naturally ascending.
	sortDesc := sortKey == paymentusecase.DefaultSortKey
	switch strings.TrimSpace(c.Query("order")) {
	case "":
	case "asc":
		sortDesc = false
	case "desc":
		sortDesc = true
	default:
		return paymentusecase.ListFilters{}, apierror.BadRequest("order must be asc or desc")
	}

	return paymentusecase.ListFilters{
		InvoiceID:     invoiceID,
		ContractID:    contractID,
		RoomID:        roomID,
		DormitoryID:   dormitoryID,
		TenantID:      tenantID,
		DateFrom:      dateFrom,
		DateTo:        dateTo,
		Search:        strings.TrimSpace(c.Query("q")),
		Columns:       columns,
		PaymentMethod: paymentMethod,
		SortKey:       sortKey,
		SortDesc:      sortDesc,
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
// @Param q query string false "Filter by tenant name, room number, dormitory name or reference no."
// @Param f[column] query string false "Per-column substring filter, e.g. f[room_number]=101; column must be one of tenant_name, room_number, dormitory_name, reference_no"
// @Param payment_method query string false "Filter by payment method: cash, transfer, credit_card, other"
// @Param sort query string false "Sort field: tenant_name, room_number, amount, payment_date (default payment_date)"
// @Param order query string false "Sort direction: asc or desc"
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
// @Summary Record a payment against an invoice, split across one or more payment methods
// @Description Records a payment for an invoice as one or more line items (e.g. part cash, part transfer), each with its own method, amount and optional reference number. Once the invoice's recorded payments reach its total_amount, the invoice is automatically marked paid.
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

	items := make([]paymentusecase.ItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, paymentusecase.ItemInput{
			PaymentMethod: item.PaymentMethod,
			Amount:        item.Amount,
			ReferenceNo:   item.ReferenceNo,
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	payment, err := h.usecase.Create(ctx, paymentusecase.CreateInput{
		InvoiceID:   req.InvoiceID,
		PaymentDate: req.PaymentDate,
		Note:        req.Note,
		Items:       items,
		CreatedBy:   &requesterID,
	})
	if err != nil {
		if errors.Is(err, paymentdomain.ErrRequiredPaymentData) {
			return apierror.BadRequest("invoice_id and payment_date are required")
		}
		if errors.Is(err, paymentdomain.ErrRequiredItems) {
			return apierror.BadRequest("at least one payment item is required")
		}
		if errors.Is(err, paymentdomain.ErrInvalidAmount) {
			return apierror.BadRequest("each payment item's amount must be greater than zero")
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
// @Description Deletes a payment (and its items) and re-evaluates the invoice's paid status accordingly.
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
