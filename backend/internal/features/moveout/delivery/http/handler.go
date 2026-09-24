package http

import (
	"context"
	"errors"
	"time"

	moveoutdomain "apihorpug/internal/features/moveout/domain"
	moveoutusecase "apihorpug/internal/features/moveout/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase *moveoutusecase.Service
}

func NewHandler(usecase *moveoutusecase.Service) *Handler {
	return &Handler{usecase: usecase}
}

type deductionRequest struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

type moveOutRequest struct {
	MoveOutDate time.Time          `json:"move_out_date"`
	Note        string             `json:"note"`
	Deductions  []deductionRequest `json:"deductions"`
}

type presetRequest struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type replacePresetsRequest struct {
	Presets []presetRequest `json:"presets"`
}

func (r moveOutRequest) manualItems() []moveoutusecase.ManualItem {
	items := make([]moveoutusecase.ManualItem, 0, len(r.Deductions))
	for _, d := range r.Deductions {
		items = append(items, moveoutusecase.ManualItem{Description: d.Description, Amount: d.Amount})
	}
	return items
}

func moveOutError(err error, fallback string) error {
	switch {
	case errors.Is(err, moveoutdomain.ErrContractNotFound):
		return apierror.NotFound("contract not found")
	case errors.Is(err, moveoutdomain.ErrMoveOutNotFound):
		return apierror.NotFound("move-out not found")
	case errors.Is(err, moveoutdomain.ErrDormitoryNotFound):
		return apierror.NotFound("dormitory not found")
	case errors.Is(err, moveoutdomain.ErrContractNotActive):
		return apierror.Conflict(err.Error()).WithSlug("contract_not_active")
	case errors.Is(err, moveoutdomain.ErrAlreadyMovedOut):
		return apierror.Conflict(err.Error()).WithSlug("already_moved_out")
	case errors.Is(err, moveoutdomain.ErrRequiredDate),
		errors.Is(err, moveoutdomain.ErrDateBeforeStart),
		errors.Is(err, moveoutdomain.ErrInvalidItem),
		errors.Is(err, moveoutdomain.ErrInvalidPreset):
		return apierror.BadRequest(err.Error())
	}
	return apierror.Internal(fallback)
}

// Preview godoc
// @Summary Preview a contract's move-out settlement
// @Description Works out the deposit reckoning for moving out on move_out_date with the given deductions, without saving anything: unpaid invoice balances, rent for the days stayed in the move-out month (or a credit if that month was invoiced in full), unbilled meter readings, and the deductions; then the refund or the amount still due.
// @Tags move-outs
// @Accept json
// @Produce json
// @Param id path string true "Contract ID"
// @Param request body moveOutRequest true "Move-out date and deductions"
// @Success 200 {object} moveoutdomain.Preview
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Security BearerAuth
// @Router /contracts/{id}/move-out/preview [post]
func (h *Handler) Preview(c fiber.Ctx) error {
	contractID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid contract id")
	}
	var req moveOutRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	preview, err := h.usecase.Preview(ctx, contractID, requesterID, req.MoveOutDate, req.manualItems())
	if err != nil {
		return moveOutError(err, "failed to preview move-out")
	}
	return apiresponse.OK(c, preview)
}

// Confirm godoc
// @Summary Confirm a contract's move-out
// @Description Settles the move-out: records the settlement, pays the contract's unpaid invoices from the deposit (oldest first, each as a "deposit" payment with its own receipt number), terminates the contract and frees the room. Cannot be undone.
// @Tags move-outs
// @Accept json
// @Produce json
// @Param id path string true "Contract ID"
// @Param request body moveOutRequest true "Move-out date, note and deductions"
// @Success 201 {object} moveoutdomain.MoveOut
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 409 {object} apierror.Error
// @Security BearerAuth
// @Router /contracts/{id}/move-out [post]
func (h *Handler) Confirm(c fiber.Ctx) error {
	contractID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid contract id")
	}
	var req moveOutRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 20*time.Second)
	defer cancel()

	moveOut, err := h.usecase.Confirm(ctx, contractID, requesterID, req.MoveOutDate, req.Note, req.manualItems(), c.IP())
	if err != nil {
		return moveOutError(err, "failed to confirm move-out")
	}
	return apiresponse.Created(c, moveOut)
}

// Get godoc
// @Summary Get a confirmed move-out
// @Tags move-outs
// @Produce json
// @Param id path string true "Move-out ID"
// @Success 200 {object} moveoutdomain.MoveOut
// @Failure 404 {object} apierror.Error
// @Security BearerAuth
// @Router /move-outs/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid move-out id")
	}
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	moveOut, err := h.usecase.GetByID(ctx, id, requesterID)
	if err != nil {
		return moveOutError(err, "failed to load move-out")
	}
	return apiresponse.OK(c, moveOut)
}

// ListPresets godoc
// @Summary List a dormitory's ready-made move-out deductions
// @Tags move-outs
// @Produce json
// @Param id path string true "Dormitory ID"
// @Success 200 {array} moveoutdomain.Preset
// @Failure 404 {object} apierror.Error
// @Security BearerAuth
// @Router /dormitories/{id}/deduction-presets [get]
func (h *Handler) ListPresets(c fiber.Ctx) error {
	dormitoryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid dormitory id")
	}
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	presets, err := h.usecase.ListPresets(ctx, dormitoryID, requesterID)
	if err != nil {
		return moveOutError(err, "failed to load deduction presets")
	}
	return apiresponse.OK(c, presets)
}

// ReplacePresets godoc
// @Summary Replace a dormitory's ready-made move-out deductions
// @Tags move-outs
// @Accept json
// @Produce json
// @Param id path string true "Dormitory ID"
// @Param request body replacePresetsRequest true "The full list, in display order"
// @Success 200 {array} moveoutdomain.Preset
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Security BearerAuth
// @Router /dormitories/{id}/deduction-presets [put]
func (h *Handler) ReplacePresets(c fiber.Ctx) error {
	dormitoryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid dormitory id")
	}
	var req replacePresetsRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	presets := make([]moveoutdomain.Preset, 0, len(req.Presets))
	for _, p := range req.Presets {
		presets = append(presets, moveoutdomain.Preset{Name: p.Name, Amount: p.Amount})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	saved, err := h.usecase.ReplacePresets(ctx, dormitoryID, requesterID, presets)
	if err != nil {
		return moveOutError(err, "failed to save deduction presets")
	}
	return apiresponse.OK(c, saved)
}
