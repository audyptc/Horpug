package v1

import (
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	auth *usecase.AuthUseCase
}

func NewAuthHandler(auth *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Login godoc
// @Summary      เข้าสู่ระบบ
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body domain.LoginRequest true "Email และ Password"
// @Success      200  {object}  domain.LoginResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": "invalid request body",
		})
	}

	resp, err := h.auth.Login(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": resp})
}

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req domain.RefreshRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": "invalid request body",
		})
	}

	resp, err := h.auth.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": resp})
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req domain.LogoutRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": "invalid request body",
		})
	}
	_ = h.auth.Logout(c.Context(), req.RefreshToken)
	return c.JSON(fiber.Map{"success": true, "message": "logged out successfully"})
}
