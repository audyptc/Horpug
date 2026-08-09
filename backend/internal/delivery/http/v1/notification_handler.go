package v1

import (
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type NotificationHandler struct {
	notifs *usecase.NotificationUseCase
}

func NewNotificationHandler(notifs *usecase.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{notifs: notifs}
}

func (h *NotificationHandler) List(c fiber.Ctx) error {
	items, err := h.notifs.List(c.Context())
	if err != nil {
		return err
	}
	return response.OK(c, items)
}
