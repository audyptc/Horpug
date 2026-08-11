package apiresponse

import "github.com/gofiber/fiber/v3"

func OK(c fiber.Ctx, data any) error {
	return c.JSON(data)
}

func Created(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(data)
}

func Message(c fiber.Ctx, message string) error {
	return c.JSON(fiber.Map{"message": message})
}
