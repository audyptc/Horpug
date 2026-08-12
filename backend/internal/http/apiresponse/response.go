package apiresponse

import "github.com/gofiber/fiber/v3"

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func OK(c fiber.Ctx, data any) error {
	return c.JSON(data)
}

func Paginated(c fiber.Ctx, data any, page, perPage int, total int64) error {
	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}

	return c.JSON(fiber.Map{
		"data": data,
		"meta": Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

func Created(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(data)
}

func Message(c fiber.Ctx, message string) error {
	return c.JSON(fiber.Map{"message": message})
}
