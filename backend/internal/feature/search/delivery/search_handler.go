package delivery

import (
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/feature/search/usecase"

	"github.com/gofiber/fiber/v3"
)

type SearchHandler struct {
	search *usecase.SearchUseCase
}

func NewSearchHandler(search *usecase.SearchUseCase) *SearchHandler {
	return &SearchHandler{search: search}
}

func (h *SearchHandler) Global(c fiber.Ctx) error {
	q := c.Query("q")
	dormitoryID, _ := c.Locals("dormitory_id").(string)
	results, err := h.search.Search(c.Context(), dormitoryID, q)
	if err != nil {
		return err
	}
	return response.OK(c, results)
}
