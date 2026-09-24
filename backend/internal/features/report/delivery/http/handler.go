package http

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	reportdomain "apihorpug/internal/features/report/domain"
	reportusecase "apihorpug/internal/features/report/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *reportusecase.Service
}

func NewHandler(usecase *reportusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

type monthlyParams struct {
	requesterID uuid.UUID
	dormitoryID *uuid.UUID
	year, month int
}

func parseMonthly(c fiber.Ctx) (monthlyParams, error) {
	var p monthlyParams
	var ok bool
	if p.requesterID, ok = middleware.UserID(c); !ok {
		return p, apierror.Unauthorized("authentication required")
	}
	var err error
	if p.year, err = strconv.Atoi(strings.TrimSpace(c.Query("year"))); err != nil {
		return p, apierror.BadRequest("invalid year")
	}
	if p.month, err = strconv.Atoi(strings.TrimSpace(c.Query("month"))); err != nil {
		return p, apierror.BadRequest("invalid month")
	}
	if raw := strings.TrimSpace(c.Query("dormitory_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return p, apierror.BadRequest("invalid dormitory_id")
		}
		p.dormitoryID = &id
	}
	return p, nil
}

func reportError(err error, fallback string) error {
	if errors.Is(err, reportusecase.ErrInvalidPeriod) {
		return apierror.BadRequest(err.Error())
	}
	return apierror.Internal(fallback)
}

// Monthly godoc
// @Summary Monthly financial report
// @Description Income received in the month (by payment method), the month's invoices (billed, collected, outstanding, by item type), expenses by category, net income, arrears across all periods, current occupancy, and a per-dormitory breakdown, over the dormitories the requester manages or one dormitory.
// @Tags reports
// @Produce json
// @Param year query int true "Year (AD)"
// @Param month query int true "Month (1-12)"
// @Param dormitory_id query string false "Limit to one dormitory"
// @Success 200 {object} reportdomain.MonthlyReport
// @Failure 400 {object} apierror.Error
// @Security BearerAuth
// @Router /reports/monthly [get]
func (h *Handler) Monthly(c fiber.Ctx) error {
	p, err := parseMonthly(c)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
	defer cancel()

	var rep reportdomain.MonthlyReport
	if rep, err = h.usecase.Monthly(ctx, p.requesterID, p.dormitoryID, p.year, p.month); err != nil {
		return reportError(err, "failed to build report")
	}
	return apiresponse.OK(c, rep)
}

// Export godoc
// @Summary Download the monthly report as an Excel workbook
// @Description The monthly summary plus sheets listing the month's payments (voided ones marked), the period's invoices, and the month's expenses.
// @Tags reports
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param year query int true "Year (AD)"
// @Param month query int true "Month (1-12)"
// @Param dormitory_id query string false "Limit to one dormitory"
// @Success 200 {file} file
// @Failure 400 {object} apierror.Error
// @Security BearerAuth
// @Router /reports/monthly/export [get]
func (h *Handler) Export(c fiber.Ctx) error {
	p, err := parseMonthly(c)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	data, err := h.usecase.Export(ctx, p.requesterID, p.dormitoryID, p.year, p.month)
	if err != nil {
		return reportError(err, "failed to export report")
	}

	c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="report-%04d-%02d.xlsx"`, p.year, p.month))
	return c.Send(data)
}
