package v1

import (
	"strconv"

	"apigofiberhorpug/internal/delivery/http/apierror"

	"github.com/gofiber/fiber/v3"
)

const (
	defaultPage    = 1
	defaultPerPage = 10
	maxPerPage     = 100
)

func parsePaginationQuery(c fiber.Ctx) (page, perPage, offset int, err error) {
	page = defaultPage
	perPage = defaultPerPage

	if rawPage := c.Query("page"); rawPage != "" {
		page, err = strconv.Atoi(rawPage)
		if err != nil || page < 1 {
			return 0, 0, 0, apierror.BadRequest("page must be a positive integer")
		}
	}

	if rawPerPage := c.Query("per_page"); rawPerPage != "" {
		perPage, err = strconv.Atoi(rawPerPage)
		if err != nil || perPage < 1 {
			return 0, 0, 0, apierror.BadRequest("per_page must be a positive integer")
		}
	}

	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	offset = (page - 1) * perPage
	return page, perPage, offset, nil
}
