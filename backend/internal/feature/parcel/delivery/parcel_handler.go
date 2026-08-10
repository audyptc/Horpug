package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/parcel/domain"
	"apigofiberhorpug/internal/feature/parcel/usecase"

	"github.com/gofiber/fiber/v3"
)

type ParcelHandler struct {
	parcels *usecase.ParcelUseCase
}

func NewParcelHandler(parcels *usecase.ParcelUseCase) *ParcelHandler {
	return &ParcelHandler{parcels: parcels}
}

func (h *ParcelHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	list, total, err := h.parcels.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *ParcelHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.parcels.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

func (h *ParcelHandler) Create(c fiber.Ctx) error {
	var req domain.CreateParcelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateParcelRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.parcels.Create(c.Context(), dormitoryID, &req)
	if err != nil {
		return err
	}
	return response.Created(c, p)
}

func (h *ParcelHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateParcelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateUpdateParcelRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.parcels.Update(c.Context(), dormitoryID, c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

func (h *ParcelHandler) Delete(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.parcels.Delete(c.Context(), dormitoryID, c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "parcel deleted")
}
