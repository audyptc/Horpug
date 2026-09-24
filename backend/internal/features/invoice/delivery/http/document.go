package http

import (
	"context"
	"errors"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Document godoc
// @Summary Get an invoice as a printable invoice/receipt
// @Description Returns the invoice with its items, the issuing dormitory, recorded payments, the paid and outstanding amounts, and (while money is owed and the dormitory has a PromptPay ID) a PromptPay QR payload for the outstanding amount.
// @Tags invoices
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} invoicedomain.Document
// @Failure 400 {object} apierror.Error
// @Failure 404 {object} apierror.Error
// @Failure 500 {object} apierror.Error
// @Security BearerAuth
// @Router /invoices/{id}/document [get]
func (h *Handler) Document(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.BadRequest("invalid invoice id")
	}

	requesterID, ok := middleware.UserID(c)
	if !ok {
		return apierror.Unauthorized("authentication required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	doc, err := h.usecase.GetDocument(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, invoicedomain.ErrInvoiceNotFound) {
			return apierror.NotFound("invoice not found")
		}
		return apierror.Internal("failed to load invoice document")
	}

	return apiresponse.OK(c, doc)
}
