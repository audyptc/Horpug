package delivery

import (
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/report/usecase"

	"github.com/gofiber/fiber/v3"
)

type ReportHandler struct {
	reports *usecase.ReportUseCase
}

func NewReportHandler(reports *usecase.ReportUseCase) *ReportHandler {
	return &ReportHandler{reports: reports}
}

func (h *ReportHandler) Income(c fiber.Ctx) error {
	from := c.Query("from")
	to := c.Query("to")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	data, err := h.reports.IncomeReport(c.Context(), dormitoryID, from, to)
	if err != nil {
		return err
	}
	return response.OK(c, data)
}

func (h *ReportHandler) Expenses(c fiber.Ctx) error {
	from := c.Query("from")
	to := c.Query("to")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	data, err := h.reports.ExpenseReport(c.Context(), dormitoryID, from, to)
	if err != nil {
		return err
	}
	return response.OK(c, data)
}

func (h *ReportHandler) Occupancy(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	data, err := h.reports.OccupancyReport(c.Context(), dormitoryID)
	if err != nil {
		return err
	}
	return response.OK(c, data)
}
