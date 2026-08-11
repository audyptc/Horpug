package http

import (
	"errors"

	"apigofiberhorpug/internal/shared/http/apierror"

	"github.com/gofiber/fiber/v3"
)

func ErrorHandler(c fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	message := "internal server error"

	var appErr *apierror.AppError
	if errors.As(err, &appErr) {
		status = appErr.Code
		message = appErr.Message
		if message == "" {
			message = err.Error()
		}
	} else {
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			status = fiberErr.Code
			message = fiberErr.Message
		}
	}

	if message == "" {
		message = fiber.ErrInternalServerError.Message
	}

	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}
