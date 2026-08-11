package delivery

import (
	"apigofiberhorpug/internal/feature/parcel/domain"
	"apigofiberhorpug/internal/feature/parcel/usecase"
	"apigofiberhorpug/internal/shared/http/apierror"
	"apigofiberhorpug/internal/shared/http/httputil"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type ParcelHandler struct {
	parcels *usecase.ParcelUseCase
}

func NewParcelHandler(parcels *usecase.ParcelUseCase) *ParcelHandler {
	return &ParcelHandler{parcels: parcels}
}

// List godoc
// @Summary      รายการพัสดุ
// @Tags         parcels
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.Parcel
// @Router       /parcels [get]
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

// GetByID godoc
// @Summary      ดูข้อมูลพัสดุตาม ID
// @Tags         parcels
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Parcel ID"
// @Success      200  {object}  domain.Parcel
// @Router       /parcels/{id} [get]
func (h *ParcelHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.parcels.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

// Create godoc
// @Summary      สร้างรายการพัสดุ
// @Tags         parcels
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body domain.CreateParcelRequest true "Parcel payload"
// @Success      201  {object}  domain.Parcel
// @Router       /parcels [post]
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

// Update godoc
// @Summary      แก้ไขรายการพัสดุ
// @Tags         parcels
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Parcel ID"
// @Param        body body domain.UpdateParcelRequest true "Parcel payload"
// @Success      200  {object}  domain.Parcel
// @Router       /parcels/{id} [put]
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

// Delete godoc
// @Summary      ลบรายการพัสดุ
// @Tags         parcels
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Parcel ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /parcels/{id} [delete]
func (h *ParcelHandler) Delete(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.parcels.Delete(c.Context(), dormitoryID, c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "parcel deleted")
}
