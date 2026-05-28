package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"
	"apigofiberhorpug/internal/validator"

	"github.com/gofiber/fiber/v3"
)

type ContractHandler struct {
	contracts   *usecase.ContractUseCase
	activityLog *usecase.ActivityLogUseCase
}

func NewContractHandler(contracts *usecase.ContractUseCase, activityLog *usecase.ActivityLogUseCase) *ContractHandler {
	return &ContractHandler{contracts: contracts, activityLog: activityLog}
}

func (h *ContractHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	list, total, err := h.contracts.List(c.Context(), perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *ContractHandler) GetByID(c fiber.Ctx) error {
	d, err := h.contracts.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, d)
}

func (h *ContractHandler) Create(c fiber.Ctx) error {
	var req domain.CreateContractRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validator.CreateContractRequest(&req); err != nil {
		return err
	}
	d, err := h.contracts.Create(c.Context(), &req)
	if err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, domain.ActivityCreate, "contract", d.ID, d)
	return response.Created(c, d)
}

func (h *ContractHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateContractRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	d, err := h.contracts.Update(c.Context(), c.Params("id"), &req)
	if err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, domain.ActivityUpdate, "contract", d.ID, d)
	return response.OK(c, d)
}

func (h *ContractHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.contracts.Delete(c.Context(), id); err != nil {
		return err
	}
	actorID, _ := c.Locals("user_id").(string)
	h.activityLog.Log(c.Context(), actorID, domain.ActivityDelete, "contract", id, nil)
	return response.Message(c, "contract deleted")
}
