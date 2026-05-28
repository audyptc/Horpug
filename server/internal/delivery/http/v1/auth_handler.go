package v1

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/delivery/http/response"
	"apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/usecase"
	"apigofiberhorpug/internal/validator"

	"github.com/gofiber/fiber/v3"
)

const refreshTokenCookieName = "refresh_token"

type AuthHandler struct {
	auth *usecase.AuthUseCase
}

func NewAuthHandler(auth *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) setRefreshCookie(c fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     "/api/v1/auth",
		MaxAge:   h.auth.RefreshMaxAge(),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})
}

func clearRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})
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
	if err := validator.LoginRequest(&req); err != nil {
		return err
	}
	resp, err := h.auth.Login(c.Context(), &req)
	if err != nil {
		return err
	}
	h.setRefreshCookie(c, resp.RefreshToken)
	resp.RefreshToken = ""
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
	rawToken := c.Cookies(refreshTokenCookieName)
	if rawToken == "" {
		return apierror.Unauthorized("missing refresh token")
	}
	resp, err := h.auth.Refresh(c.Context(), rawToken)
	if err != nil {
		clearRefreshCookie(c)
		return err
	}
	h.setRefreshCookie(c, resp.RefreshToken)
	resp.RefreshToken = ""
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
	rawToken := c.Cookies(refreshTokenCookieName)
	if rawToken != "" {
		_ = h.auth.Logout(c.Context(), rawToken)
	}
	clearRefreshCookie(c)
	return response.Message(c, "logged out successfully")
}
