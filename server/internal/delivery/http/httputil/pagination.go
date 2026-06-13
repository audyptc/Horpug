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

	// รองรับทั้ง per_page และ limit
	if rawPerPage := c.Query("per_page"); rawPerPage != "" {
		perPage, err = strconv.Atoi(rawPerPage)
		if err != nil || perPage < 1 {
			return 0, 0, 0, apierror.BadRequest("per_page must be a positive integer")
		}
	} else if rawLimit := c.Query("limit"); rawLimit != "" {
		perPage, err = strconv.Atoi(rawLimit)
		if err != nil || perPage < 1 {
			return 0, 0, 0, apierror.BadRequest("limit must be a positive integer")
		}
	}

	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	// รองรับทั้ง offset โดยตรง และคำนวณจาก page
	if rawOffset := c.Query("offset"); rawOffset != "" {
		offset, err = strconv.Atoi(rawOffset)
		if err != nil || offset < 0 {
			return 0, 0, 0, apierror.BadRequest("offset must be a non-negative integer")
		}
		if perPage > 0 {
			page = offset/perPage + 1
		}
	} else {
		offset = (page - 1) * perPage
	}

	return page, perPage, offset, nil
}
