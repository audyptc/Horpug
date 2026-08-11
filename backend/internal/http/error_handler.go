package http

import (
	"errors"

	"apihorpug/internal/http/apierror"

	"github.com/gofiber/fiber/v3"
)

func ErrorHandler(c fiber.Ctx, err error) error {
	var appErr *apierror.Error
	if errors.As(err, &appErr) {
		return c.Status(appErr.Code).JSON(fiber.Map{"error": appErr.Message})
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
}
