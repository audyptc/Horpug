package http

import (
	"context"
	"errors"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"
	invoiceusecase "apihorpug/internal/features/invoice/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type generateInvoicesRequest struct {
	DormitoryID uuid.UUID   `json:"dormitory_id"`
	PeriodYear  int         `json:"period_year"`
	PeriodMonth int         `json:"period_month"`
	IssueDate   time.Time   `json:"issue_date"`
	DueDate     time.Time   `json:"due_date"`
	Note        string      `json:"note"`
	ContractIDs []uuid.UUID `json:"contract_ids"`
}

func generationError(err error) error {
	switch {
	case errors.Is(err, invoicedomain.ErrRequiredGenerateData):
		return apierror.BadRequest("dormitory_id, period_year, period_month, issue_date and due_date are required")
	case errors.Is(err, invoicedomain.ErrInvalidInvoicePeriod):
		return apierror.BadRequest("period_month must be between 1 and 12")
	case errors.Is(err, invoicedomain.ErrInvalidInvoiceDates):
		return apierror.BadRequest("due_date must not be before issue_date")
	case errors.Is(err, invoicedomain.ErrDormitoryNotFound):
		return apierror.NotFound("dormitory not found")
	}
	return nil
}

// PreviewGeneration godoc
// @Summary Preview bulk invoice generation for a dormitory and period
// @Description Lists the dormitory's active contracts that had started by the end of the period, flagging those already invoiced for it, those missing an electricity or water reading in the period, and those whose end date has passed.
// @Tags invoices
// @Produce json
// @Param dormitory_id query string true "Dormitory ID"
// @Param period_year query int true "Billing year"
// @Param period_month query int true "Billing month (1-12)"
// @Success 200 {array} invoicedomain.GenerationCandidate
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /invoices/generate/preview [get]
func (h *Handler) PreviewGeneration(c fiber.Ctx) error {
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	dormitoryID, err := parseUUIDQuery(c, "dormitory_id")
	if err != nil {
		return err
	}
	periodYear, err := parseIntQuery(c, "period_year")
	if err != nil {
		return err
	}
	periodMonth, err := parseIntQuery(c, "period_month")
	if err != nil {
		return err
	}
	if dormitoryID == nil || periodYear == nil || periodMonth == nil {
		return apierror.BadRequest("dormitory_id, period_year and period_month are required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	candidates, err := h.usecase.PreviewGeneration(ctx, requesterID, *dormitoryID, *periodYear, *periodMonth)
	if err != nil {
		if apiErr := generationError(err); apiErr != nil {
			return apiErr
		}
		return apierror.Internal("failed to preview invoice generation")
	}

	return apiresponse.OK(c, candidates)
}

// Generate godoc
// @Summary Generate a period's invoices for a dormitory in one go
// @Description Creates the period's invoice (rent plus the period's meter readings) for each active contract in the dormitory that has none yet, or only for contract_ids when given. Each invoice is created independently: the response lists those created, the number skipped as already invoiced, and any that failed.
// @Tags invoices
// @Accept json
// @Produce json
// @Param request body generateInvoicesRequest true "Generation payload"
// @Success 200 {object} invoicedomain.GenerationResult
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /invoices/generate [post]
func (h *Handler) Generate(c fiber.Ctx) error {
	var req generateInvoicesRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	// One transaction per invoice, so a large dormitory needs more than the
	// usual per-request budget.
	ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
	defer cancel()

	result, err := h.usecase.Generate(ctx, invoiceusecase.GenerateInput{
		DormitoryID: req.DormitoryID,
		PeriodYear:  req.PeriodYear,
		PeriodMonth: req.PeriodMonth,
		IssueDate:   req.IssueDate,
		DueDate:     req.DueDate,
		Note:        req.Note,
		ContractIDs: req.ContractIDs,
		CreatedBy:   requesterID,
	}, c.IP())
	if err != nil {
		if apiErr := generationError(err); apiErr != nil {
			return apiErr
		}
		return apierror.Internal("failed to generate invoices")
	}

	return apiresponse.OK(c, result)
}
