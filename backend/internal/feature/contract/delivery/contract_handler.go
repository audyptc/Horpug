package delivery

import (
	aldomain "apigofiberhorpug/internal/feature/activitylog/domain"
	alusecase "apigofiberhorpug/internal/feature/activitylog/usecase"
	"apigofiberhorpug/internal/feature/contract/domain"
	"apigofiberhorpug/internal/feature/contract/usecase"
	"apigofiberhorpug/internal/shared/http/apierror"
	"apigofiberhorpug/internal/shared/http/httputil"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type ContractHandler struct {
	contracts   *usecase.ContractUseCase
	activityLog *alusecase.ActivityLogUseCase
}

func NewContractHandler(contracts *usecase.ContractUseCase, activityLog *alusecase.ActivityLogUseCase) *ContractHandler {
	return &ContractHandler{contracts: contracts, activityLog: activityLog}
}

// List godoc
// @Summary      รายชื่อสัญญาเช่า
// @Tags         contracts
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Contract
// @Router       /contracts [get]
func (h *ContractHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	list, total, err := h.contracts.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

// GetByID godoc
// @Summary      ดูข้อมูลสัญญาเช่าตาม ID
// @Tags         contracts
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Contract ID"
// @Success      200  {object}  domain.ContractDetail
// @Router       /contracts/{id} [get]
func (h *ContractHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.contracts.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

// Create godoc
// @Summary      สร้างสัญญาเช่า
// @Tags         contracts
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateContractRequest true "Contract payload"
// @Success      201  {object}  domain.Contract
// @Router       /contracts [post]
func (h *ContractHandler) Create(c fiber.Ctx) error {
	var req domain.CreateContractRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateContractRequest(&req); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.contracts.Create(c.Context(), dormitoryID, &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityCreate, "contract", d.ID, d)
	return response.Created(c, d)
}

// Update godoc
// @Summary      แก้ไขสัญญาเช่า
// @Tags         contracts
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Contract ID"
// @Param        body body domain.UpdateContractRequest true "Contract payload"
// @Success      200  {object}  domain.Contract
// @Router       /contracts/{id} [put]
func (h *ContractHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateContractRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	actorID, _ := c.Locals("user_id").(string)
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	d, err := h.contracts.Update(c.Context(), dormitoryID, c.Params("id"), &req, actorID)
	if err != nil {
		return err
	}
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityUpdate, "contract", d.ID, d)
	return response.OK(c, d)
}

// Delete godoc
// @Summary      ลบสัญญาเช่า
// @Tags         contracts
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Contract ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /contracts/{id} [delete]
func (h *ContractHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.contracts.Delete(c.Context(), dormitoryID, id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.LogForDormitory(c.Context(), actorID, dormitoryID, aldomain.ActivityDelete, "contract", id, nil)
	return response.Message(c, "contract deleted")
}
