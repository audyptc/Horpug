package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/parking/domain"
	"apigofiberhorpug/internal/feature/parking/usecase"

	"github.com/gofiber/fiber/v3"
)

type ParkingHandler struct {
	parking *usecase.ParkingUseCase
}

func NewParkingHandler(parking *usecase.ParkingUseCase) *ParkingHandler {
	return &ParkingHandler{parking: parking}
}

func (h *ParkingHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	list, total, err := h.parking.List(c.Context(), dormitoryID, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}

func (h *ParkingHandler) GetByID(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.parking.GetByID(c.Context(), dormitoryID, c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

func (h *ParkingHandler) Create(c fiber.Ctx) error {
	var req domain.CreateParkingSlotRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateCreateParkingSlotRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.parking.Create(c.Context(), dormitoryID, &req)
	if err != nil {
		return err
	}
	return response.Created(c, p)
}

func (h *ParkingHandler) Update(c fiber.Ctx) error {
	var req domain.UpdateParkingSlotRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	if err := validateUpdateParkingSlotRequest(&req); err != nil {
		return err
	}
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	p, err := h.parking.Update(c.Context(), dormitoryID, c.Params("id"), &req)
	if err != nil {
		return err
	}
	return response.OK(c, p)
}

func (h *ParkingHandler) Delete(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	if err := h.parking.Delete(c.Context(), dormitoryID, c.Params("id")); err != nil {
		return err
	}
	return response.Message(c, "parking slot deleted")
}
