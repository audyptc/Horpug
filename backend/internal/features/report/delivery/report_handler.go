package delivery

import (
	"apigofiberhorpug/internal/features/report/domain"
	"apigofiberhorpug/internal/features/report/usecase"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type ReportHandler struct {
	reports *usecase.ReportUseCase
}

func NewReportHandler(reports *usecase.ReportUseCase) *ReportHandler {
	return &ReportHandler{reports: reports}
}

// Income godoc
// @Summary      รายงานรายรับ
// @Tags         reports
// @Security     ApiKeyAuth
// @Produce      json
// @Param        from query string false "From date (YYYY-MM-DD)"
// @Param        to query string false "To date (YYYY-MM-DD)"
// @Success      200  {object}  domain.IncomeReport
// @Router       /reports/income [get]
func (h *ReportHandler) Income(c fiber.Ctx) error {
	from := c.Query("from")
	to := c.Query("to")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	var data *domain.IncomeReport
	data, err := h.reports.IncomeReport(c.Context(), dormitoryID, from, to)
	if err != nil {
		return err
	}
	return response.OK(c, data)
}

// Expenses godoc
// @Summary      รายงานรายจ่าย
// @Tags         reports
// @Security     ApiKeyAuth
// @Produce      json
// @Param        from query string false "From date (YYYY-MM-DD)"
// @Param        to query string false "To date (YYYY-MM-DD)"
// @Success      200  {object}  domain.ExpenseReport
// @Router       /reports/expenses [get]
func (h *ReportHandler) Expenses(c fiber.Ctx) error {
	from := c.Query("from")
	to := c.Query("to")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	var data *domain.ExpenseReport
	data, err := h.reports.ExpenseReport(c.Context(), dormitoryID, from, to)
	if err != nil {
		return err
	}
	return response.OK(c, data)
}

// Occupancy godoc
// @Summary      รายงานอัตราการเข้าพัก
// @Tags         reports
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {object}  domain.OccupancyReport
// @Router       /reports/occupancy [get]
func (h *ReportHandler) Occupancy(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	var data *domain.OccupancyReport
	data, err := h.reports.OccupancyReport(c.Context(), dormitoryID)
	if err != nil {
		return err
	}
	return response.OK(c, data)
}
