package delivery

import (
	aldomain "apigofiberhorpug/internal/feature/activitylog/domain"
	alusecase "apigofiberhorpug/internal/feature/activitylog/usecase"
	"apigofiberhorpug/internal/feature/bill/domain"
	"apigofiberhorpug/internal/feature/bill/usecase"
	"apigofiberhorpug/internal/shared/http/apierror"
	"apigofiberhorpug/internal/shared/http/httputil"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type BillHandler struct {
	bills       *usecase.BillUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewBillHandler(bills *usecase.BillUseCase, activityLog *alusecase.ActivityLogUseCase) *BillHandler {
	return &BillHandler{bills: bills, activityLog: activityLog}
}

// List godoc
// @Summary      รายการบิล
// @Tags         bills
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Bill
// @Router       /bills [get]
func (h *BillHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	list, total, err := h.bills.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

// GetByID godoc
// @Summary      ดูข้อมูลบิลตาม ID
// @Tags         bills
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Bill ID"
// @Success      200  {object}  domain.BillDetail
// @Router       /bills/{id} [get]
func (h *BillHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.bills.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

// Create godoc
// @Summary      สร้างบิล
// @Tags         bills
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateBillRequest true "Bill payload"
// @Success      201  {object}  domain.Bill
// @Router       /bills [post]
func (h *BillHandler) Create(c fiber.Ctx) error {
	var req domain.CreateBillRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateBillRequest(&req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.bills.Create(c.Context(), dormitoryID, &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityCreate, "bill", d.ID, d)
	return response.Created(c, d)
}

// Update godoc
// @Summary      แก้ไขบิล
// @Tags         bills
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Bill ID"
// @Param        body body domain.UpdateBillRequest true "Bill payload"
// @Success      200  {object}  domain.Bill
// @Router       /bills/{id} [put]
func (h *BillHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateBillRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.bills.Update(c.Context(), dormitoryID, c.Params("id"), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityUpdate, "bill", d.ID, d)
	return response.OK(c, d)
}

// Delete godoc
// @Summary      ลบบิล
// @Tags         bills
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Bill ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /bills/{id} [delete]
func (h *BillHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.bills.Delete(c.Context(), dormitoryID, id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityDelete, "bill", id, nil)
	return response.Message(c, "bill deleted")
}
