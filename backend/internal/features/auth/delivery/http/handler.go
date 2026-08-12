package http

import (
	"context"
	"errors"
	"time"

	authdomain "apihorpug/internal/features/auth/domain"
	authusecase "apihorpug/internal/features/auth/usecase"
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/apiresponse"

	"github.com/gofiber/fiber/v3"
)

// refreshCookieName is the httpOnly cookie carrying the refresh token. It is
// never exposed to JavaScript or the JSON response body, so a token leaked
// via XSS can only ever be the short-lived access token.
const refreshCookieName = "refresh_token"

type Handler struct {
	usecase      *authusecase.Service
	cookieSecure bool
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type sessionResponse struct {
	AccessToken          string      `json:"access_token"`
	AccessTokenExpiresAt time.Time   `json:"access_token_expires_at"`
	User                 interface{} `json:"user"`
}

func NewHandler(usecase *authusecase.Service, cookieSecure bool) *Handler {
	return &Handler{usecase: usecase, cookieSecure: cookieSecure}
}

func (h *Handler) setRefreshCookie(c fiber.Ctx, token string, expiresAt time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}

func (h *Handler) clearRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}

// Login godoc
// @Summary Log in with username/email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login payload"
// @Success 200 {object} sessionResponse
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Router /auth/login [post]
func (h *Handler) Login(c fiber.Ctx) error {
	var req loginRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	if req.Username == "" || req.Password == "" {
		return apierror.BadRequest("username and password are required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	result, err := h.usecase.Login(ctx, req.Username, req.Password, c.IP())
	if err != nil {
		if errors.Is(err, authdomain.ErrInvalidCredentials) {
			return apierror.Unauthorized("invalid username or password")
		}
		if errors.Is(err, authdomain.ErrAccountInactive) {
			return apierror.Unauthorized("account is inactive")
		}
		return apierror.Internal("failed to login")
	}

	h.setRefreshCookie(c, result.RefreshToken, result.RefreshTokenExpiresAt)
	return apiresponse.OK(c, toSessionResponse(result))
}

// Refresh godoc
// @Summary Exchange the refresh token cookie for a new access/refresh token pair
// @Tags auth
// @Produce json
// @Success 200 {object} sessionResponse
// @Failure 401 {object} apierror.Error
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c fiber.Ctx) error {
	rawToken := c.Cookies(refreshCookieName)
	if rawToken == "" {
		return apierror.Unauthorized("missing refresh token")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	result, err := h.usecase.Refresh(ctx, rawToken)
	if err != nil {
		h.clearRefreshCookie(c)
		if errors.Is(err, authdomain.ErrRefreshTokenInvalid) {
			return apierror.Unauthorized("invalid or expired refresh token")
		}
		if errors.Is(err, authdomain.ErrAccountInactive) {
			return apierror.Unauthorized("account is inactive")
		}
		return apierror.Internal("failed to refresh session")
	}

	h.setRefreshCookie(c, result.RefreshToken, result.RefreshTokenExpiresAt)
	return apiresponse.OK(c, toSessionResponse(result))
}

// Logout godoc
// @Summary Revoke the refresh token cookie
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *Handler) Logout(c fiber.Ctx) error {
	rawToken := c.Cookies(refreshCookieName)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Logout(ctx, rawToken, c.IP()); err != nil {
		return apierror.Internal("failed to logout")
	}

	h.clearRefreshCookie(c)
	return apiresponse.Message(c, "logged out")
}

func toSessionResponse(result authusecase.LoginResult) sessionResponse {
	return sessionResponse{
		AccessToken:          result.AccessToken,
		AccessTokenExpiresAt: result.AccessTokenExpiresAt,
		User:                 result.User,
	}
}
