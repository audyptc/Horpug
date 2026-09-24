package http

import (
	"context"
	"errors"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"
	invoiceusecase "apihorpug/internal/features/invoice/usecase"
	"apihorpug/internal/http/apierror"

	"github.com/gofiber/fiber/v3"
)

type QRHandler struct {
	images *invoiceusecase.QRImages
}

func NewQRHandler(images *invoiceusecase.QRImages) *QRHandler {
	return &QRHandler{images: images}
}

// Image godoc
// @Summary PromptPay QR image of an invoice (public, signed link)
// @Description The link is sent in LINE overdue reminders; LINE downloads the image itself, so there is no session. The QR is for the amount still owed at request time.
// @Tags invoices
// @Produce png
// @Param token path string true "Signed invoice QR token"
// @Success 200 {file} binary
// @Failure 404 {object} apierror.Error
// @Router /public/invoice-qr/{token} [get]
func (h *QRHandler) Image(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	png, err := h.images.PNG(ctx, c.Params("token"))
	if err != nil {
		if errors.Is(err, invoicedomain.ErrInvoiceNotFound) {
			return apierror.NotFound("qr not available")
		}
		return apierror.Internal("failed to render qr")
	}

	// Short cache: the amount follows payments, but LINE may fetch the
	// image and its preview back to back.
	c.Set(fiber.HeaderCacheControl, "private, max-age=300")
	c.Set(fiber.HeaderContentType, "image/png")
	return c.Send(png)
}
