package apierror

import "github.com/gofiber/fiber/v3"

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func BadRequest(message string) *Error {
	return New(fiber.StatusBadRequest, message)
}

func NotFound(message string) *Error {
	return New(fiber.StatusNotFound, message)
}

func Unauthorized(message string) *Error {
	return New(fiber.StatusUnauthorized, message)
}

func Forbidden(message string) *Error {
	return New(fiber.StatusForbidden, message)
}

func Conflict(message string) *Error {
	return New(fiber.StatusConflict, message)
}

func Internal(message string) *Error {
	return New(fiber.StatusInternalServerError, message)
}
