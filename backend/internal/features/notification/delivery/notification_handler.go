package delivery

import (
	"apigofiberhorpug/internal/features/notification/domain"
	"apigofiberhorpug/internal/features/notification/usecase"
	"apigofiberhorpug/internal/shared/http/response"

	"github.com/gofiber/fiber/v3"
)

type NotificationHandler struct {
	notifs *usecase.NotificationUseCase
}

func NewNotificationHandler(notifs *usecase.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{notifs: notifs}
}

// List godoc
// @Summary      รายการแจ้งเตือน
// @Tags         notifications
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}  domain.NotificationItem
// @Router       /notifications [get]
func (h *NotificationHandler) List(c fiber.Ctx) error {
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	var items []*domain.NotificationItem
	items, err := h.notifs.List(c.Context(), dormitoryID)
	if err != nil {
		return err
	}
	return response.OK(c, items)
}
