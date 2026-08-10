package delivery

import (
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/search/domain"
	"apigofiberhorpug/internal/feature/search/usecase"

	"github.com/gofiber/fiber/v3"
)

type SearchHandler struct {
	search *usecase.SearchUseCase
}

func NewSearchHandler(search *usecase.SearchUseCase) *SearchHandler {
	return &SearchHandler{search: search}
}

// Global godoc
// @Summary      ค้นหาข้อมูลทั่วทั้งระบบ
// @Tags         search
// @Security     ApiKeyAuth
// @Produce      json
// @Param        q query string true "Search query"
// @Success      200  {object}  domain.SearchResults
// @Router       /search [get]
func (h *SearchHandler) Global(c fiber.Ctx) error {
	q := c.Query("q")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	var results *domain.SearchResults
	results, err := h.search.Search(c.Context(), dormitoryID, q)
	if err != nil {
		return err
	}
	return response.OK(c, results)
}
