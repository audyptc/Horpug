package http

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"
	invoiceusecase "apihorpug/internal/features/invoice/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/httputil"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *invoiceusecase.Service
}

type createInvoiceRequest struct {
	ContractID  uuid.UUID `json:"contract_id"`
	PeriodYear  int       `json:"period_year"`
	PeriodMonth int       `json:"period_month"`
	IssueDate   time.Time `json:"issue_date"`
	DueDate     time.Time `json:"due_date"`
	Note        string    `json:"note"`
}

type updateInvoiceRequest struct {
	Status  *invoicedomain.InvoiceStatus `json:"status"`
	DueDate *time.Time                   `json:"due_date"`
	Note    *string                      `json:"note"`
}

type addInvoiceItemRequest struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

func NewHandler(usecase *invoiceusecase.Service) *Handler {
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

func parseIntQuery(c fiber.Ctx, name string) (*int, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil, apierror.BadRequest("invalid " + name)
	}
	return &value, nil
}

func parseListFilters(c fiber.Ctx) (invoiceusecase.ListFilters, error) {
	contractID, err := parseUUIDQuery(c, "contract_id")
	if err != nil {
		return invoiceusecase.ListFilters{}, err
	}
	roomID, err := parseUUIDQuery(c, "room_id")
	if err != nil {
		return invoiceusecase.ListFilters{}, err
	}
	dormitoryID, err := parseUUIDQuery(c, "dormitory_id")
	if err != nil {
		return invoiceusecase.ListFilters{}, err
	}
	tenantID, err := parseUUIDQuery(c, "tenant_id")
	if err != nil {
		return invoiceusecase.ListFilters{}, err
	}
	periodYear, err := parseIntQuery(c, "period_year")
	if err != nil {
		return invoiceusecase.ListFilters{}, err
	}
	periodMonth, err := parseIntQuery(c, "period_month")
	if err != nil {
		return invoiceusecase.ListFilters{}, err
	}

	var status *invoicedomain.InvoiceStatus
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		s := invoicedomain.InvoiceStatus(raw)
		if !s.Valid() {
			return invoiceusecase.ListFilters{}, apierror.BadRequest("invalid status")
		}
		status = &s
	}

	return invoiceusecase.ListFilters{
		ContractID:  contractID,
		RoomID:      roomID,
		DormitoryID: dormitoryID,
		TenantID:    tenantID,
		Status:      status,
		PeriodYear:  periodYear,
		PeriodMonth: periodMonth,
	}, nil
}

// List godoc
// @Summary List invoices
// @Description Returns every invoice for roles with full dormitory access, otherwise only invoices under dormitories the caller manages. Optionally filter by contract, room, dormitory, tenant, status or billing period.
// @Tags invoices
// @Produce json
// @Param contract_id query string false "Filter by contract ID"
// @Param room_id query string false "Filter by room ID"
// @Param dormitory_id query string false "Filter by dormitory ID"
// @Param tenant_id query string false "Filter by tenant ID"
// @Param status query string false "Filter by status (unpaid, paid, overdue, cancelled)"
// @Param period_year query int false "Filter by billing period year"
// @Param period_month query int false "Filter by billing period month (1-12)"
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Results per page (default 10, max 100)"
// @Success 200 {object} apiresponse.Meta
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /invoices [get]
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

	invoices, total, err := h.usecase.List(ctx, requesterID, filters, perPage, offset)
	if err != nil {
		return apierror.Internal("failed to list invoices")
	}

	return apiresponse.Paginated(c, invoices, page, perPage, total)
}

// Get godoc
// @Summary Get an invoice by ID
// @Tags invoices
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} invoicedomain.Invoice
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /invoices/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid invoice id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	invoice, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, invoicedomain.ErrInvoiceNotFound) {
			return apierror.NotFound("invoice not found")
		}
		return apierror.Internal("failed to get invoice")
	}

	return apiresponse.OK(c, invoice)
}

// Create godoc
// @Summary Create an invoice for a contract's billing period
// @Description Builds an invoice from the contract's rent price plus any electricity/water meter readings recorded for the contract's room within the given period_year/period_month, each stored as a line item.
// @Tags invoices
// @Accept json
// @Produce json
// @Param request body createInvoiceRequest true "Invoice payload"
// @Success 201 {object} invoicedomain.Invoice
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /invoices [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req createInvoiceRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	invoice, err := h.usecase.Create(ctx, invoiceusecase.CreateInput{
		ContractID:  req.ContractID,
		PeriodYear:  req.PeriodYear,
		PeriodMonth: req.PeriodMonth,
		IssueDate:   req.IssueDate,
		DueDate:     req.DueDate,
		Note:        req.Note,
		CreatedBy:   &requesterID,
	})
	if err != nil {
		if errors.Is(err, invoicedomain.ErrRequiredInvoiceData) {
			return apierror.BadRequest("contract_id, period_year, period_month, issue_date and due_date are required")
		}
		if errors.Is(err, invoicedomain.ErrInvalidInvoicePeriod) {
			return apierror.BadRequest("period_month must be between 1 and 12")
		}
		if errors.Is(err, invoicedomain.ErrInvalidInvoiceDates) {
			return apierror.BadRequest("due_date must not be before issue_date")
		}
		if errors.Is(err, invoicedomain.ErrContractNotFound) {
			return apierror.NotFound("contract not found")
		}
		if errors.Is(err, invoicedomain.ErrInvoiceExists) {
			return apierror.Conflict("an invoice for this contract and period already exists")
		}
		return apierror.Internal("failed to create invoice")
	}

	return apiresponse.Created(c, invoice)
}

// Update godoc
// @Summary Update an invoice's status, due date or note
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param request body updateInvoiceRequest true "Invoice payload"
// @Success 200 {object} invoicedomain.Invoice
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /invoices/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid invoice id")
	}

	var req updateInvoiceRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	invoice, err := h.usecase.Update(ctx, id, requesterID, invoiceusecase.UpdateInput{
		Status:    req.Status,
		DueDate:   req.DueDate,
		Note:      req.Note,
		UpdatedBy: &requesterID,
	})
	if err != nil {
		if errors.Is(err, invoicedomain.ErrInvoiceNotFound) {
			return apierror.NotFound("invoice not found")
		}
		if errors.Is(err, invoicedomain.ErrInvalidInvoiceStatus) {
			return apierror.BadRequest("invalid status")
		}
		return apierror.Internal("failed to update invoice")
	}

	return apiresponse.OK(c, invoice)
}

// AddItem godoc
// @Summary Add a manual line item to an invoice
// @Description Adds an ad-hoc "other" charge to the invoice and updates its total. Not allowed once the invoice is paid or cancelled.
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param request body addInvoiceItemRequest true "Item payload"
// @Success 201 {object} invoicedomain.Invoice
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /invoices/{id}/items [post]
func (h *Handler) AddItem(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid invoice id")
	}

	var req addInvoiceItemRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	invoice, err := h.usecase.AddItem(ctx, id, requesterID, invoiceusecase.AddItemInput{
		Description: req.Description,
		Amount:      req.Amount,
		UpdatedBy:   &requesterID,
	})
	if err != nil {
		if errors.Is(err, invoicedomain.ErrInvoiceNotFound) {
			return apierror.NotFound("invoice not found")
		}
		if errors.Is(err, invoicedomain.ErrRequiredInvoiceItemData) {
			return apierror.BadRequest("description and amount are required")
		}
		if errors.Is(err, invoicedomain.ErrInvalidInvoiceItemAmount) {
			return apierror.BadRequest("amount must be greater than zero")
		}
		if errors.Is(err, invoicedomain.ErrInvoiceLocked) {
			return apierror.Conflict("invoice items cannot be changed once the invoice is paid or cancelled")
		}
		return apierror.Internal("failed to add invoice item")
	}

	return apiresponse.Created(c, invoice)
}

// RemoveItem godoc
// @Summary Remove a manually added line item from an invoice
// @Description Removes an "other" charge from the invoice and updates its total. System-generated rent/electricity/water items can't be removed this way, and neither can any item once the invoice is paid or cancelled.
// @Tags invoices
// @Produce json
// @Param id path string true "Invoice ID"
// @Param itemId path string true "Invoice item ID"
// @Success 200 {object} invoicedomain.Invoice
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /invoices/{id}/items/{itemId} [delete]
func (h *Handler) RemoveItem(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid invoice id")
	}
	itemID, err := uuid.Parse(c.Params("itemId"))
	if err != nil {
		return apierror.BadRequest("invalid invoice item id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	invoice, err := h.usecase.RemoveItem(ctx, id, itemID, requesterID)
	if err != nil {
		if errors.Is(err, invoicedomain.ErrInvoiceNotFound) {
			return apierror.NotFound("invoice not found")
		}
		if errors.Is(err, invoicedomain.ErrInvoiceItemNotFound) {
			return apierror.NotFound("invoice item not found")
		}
		if errors.Is(err, invoicedomain.ErrInvoiceItemNotRemovable) {
			return apierror.BadRequest("only manually added items can be removed")
		}
		if errors.Is(err, invoicedomain.ErrInvoiceLocked) {
			return apierror.Conflict("invoice items cannot be changed once the invoice is paid or cancelled")
		}
		return apierror.Internal("failed to remove invoice item")
	}

	return apiresponse.OK(c, invoice)
}

// SendLine godoc
// @Summary Send an invoice to the tenant via LINE
// @Description Pushes a text summary of the invoice to the tenant's linked LINE account through the dormitory's LINE Official Account. Requires the tenant to have completed the LIFF linking flow first.
// @Tags invoices
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /invoices/{id}/send-line [post]
func (h *Handler) SendLine(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid invoice id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if err := h.usecase.SendLine(ctx, id, requesterID); err != nil {
		if errors.Is(err, invoicedomain.ErrInvoiceNotFound) {
			return apierror.NotFound("invoice not found")
		}
		if errors.Is(err, invoicedomain.ErrTenantLineNotLinked) {
			return apierror.Conflict("tenant has not linked a LINE account")
		}
		if errors.Is(err, invoicedomain.ErrTenantLineUnreachable) {
			return apierror.Conflict("tenant has not added the LINE OA as a friend or has blocked it")
		}
		return apierror.Internal("failed to send invoice via LINE")
	}

	return apiresponse.Message(c, "invoice sent via LINE")
}

// Delete godoc
// @Summary Delete an invoice
// @Tags invoices
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /invoices/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid invoice id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Delete(ctx, id, requesterID); err != nil {
		if errors.Is(err, invoicedomain.ErrInvoiceNotFound) {
			return apierror.NotFound("invoice not found")
		}
		return apierror.Internal("failed to delete invoice")
	}

	return apiresponse.Message(c, "invoice deleted")
}
