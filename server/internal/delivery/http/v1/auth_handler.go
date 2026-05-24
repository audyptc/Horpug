package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/response"
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
		return apierror.BadRequest("invalid request body")
	}

	resp, err := h.auth.Login(c.Context(), &req)
	if err != nil {
		return apierror.Unauthorized(err.Error())
	}
	return response.OK(c, resp)
}

// Refresh godoc
// @Summary      ต่ออายุโทเค็น
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body domain.RefreshRequest true "Refresh Token"
// @Success      200  {object}  domain.LoginResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req domain.RefreshRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	resp, err := h.auth.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return apierror.Unauthorized(err.Error())
	}
	return response.OK(c, resp)
}

// Logout godoc
// @Summary      ออกจากระบบ
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body domain.LogoutRequest true "Refresh Token"
// @Success      200  {object}  map[string]interface{}
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req domain.LogoutRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}
	_ = h.auth.Logout(c.Context(), req.RefreshToken)
	return response.Message(c, "logged out successfully")
}
