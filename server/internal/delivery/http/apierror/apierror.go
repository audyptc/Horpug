package apierror

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

// AppError is a transport-level error for consistent HTTP responses.
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func New(code int, message string) error {
	return &AppError{Code: code, Message: message}
}

func Wrap(code int, message string, err error) error {
	if err == nil {
		return &AppError{Code: code, Message: message}
	}
	return &AppError{Code: code, Message: message, Err: err}
}

func BadRequest(message string) error {
	return New(fiber.StatusBadRequest, message)
}

func Unauthorized(message string) error {
	return New(fiber.StatusUnauthorized, message)
}

func Forbidden(message string) error {
	return New(fiber.StatusForbidden, message)
}

func NotFound(message string) error {
	return New(fiber.StatusNotFound, message)
}

func Internal(err error) error {
	return Wrap(fiber.StatusInternalServerError, "internal server error", err)
}

func Internalf(format string, args ...any) error {
	return New(fiber.StatusInternalServerError, fmt.Sprintf(format, args...))
}
