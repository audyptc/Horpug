package httputil

import (
	"strconv"

	"apigofiberhorpug/internal/delivery/http/apierror"

	"github.com/gofiber/fiber/v3"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 10
	MaxPerPage     = 100
)

func ParsePaginationQuery(c fiber.Ctx) (page, perPage, offset int, err error) {
	page = DefaultPage
	perPage = DefaultPerPage

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

	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	offset = (page - 1) * perPage
	return page, perPage, offset, nil
}
