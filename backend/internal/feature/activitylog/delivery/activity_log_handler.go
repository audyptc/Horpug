package delivery

import (
	"time"

	"apigofiberhorpug/internal/delivery/http/httputil"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/activitylog/domain"
	"apigofiberhorpug/internal/feature/activitylog/usecase"

	"github.com/gofiber/fiber/v3"
)

type ActivityLogHandler struct {
	activityLog *usecase.ActivityLogUseCase
}

func NewActivityLogHandler(activityLog *usecase.ActivityLogUseCase) *ActivityLogHandler {
	return &ActivityLogHandler{activityLog: activityLog}
}

// List godoc
// @Summary      รายการประวัติกิจกรรม
// @Tags         activity-logs
// @Security     ApiKeyAuth
// @Produce      json
// @Param        entity_type query string false "Filter by entity type"
// @Param        actor_id query string false "Filter by actor (user) ID"
// @Param        from query string false "From timestamp (RFC3339)"
// @Param        to query string false "To timestamp (RFC3339)"
// @Success      200  {array}  domain.ActivityLog
// @Router       /activity-logs [get]
func (h *ActivityLogHandler) List(c fiber.Ctx) error {
	page, perPage, offset, err := httputil.ParsePaginationQuery(c)
	if err != nil {
		return err
	}

	filter := domain.ActivityLogFilter{
		DormitoryID: c.Locals("dormitory_id", "").(string),
		EntityType:  c.Query("entity_type"),
		ActorID:     c.Query("actor_id"),
	}
	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.From = &t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.To = &t
		}
	}

	list, total, err := h.activityLog.List(c.Context(), filter, perPage, offset)
	if err != nil {
		return err
	}
	return response.Paginated(c, list, page, perPage, total)
}
