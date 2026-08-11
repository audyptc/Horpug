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

type Handler struct {
	usecase *authusecase.Service
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type sessionResponse struct {
	AccessToken           string      `json:"access_token"`
	AccessTokenExpiresAt  time.Time   `json:"access_token_expires_at"`
	RefreshToken          string      `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time   `json:"refresh_token_expires_at"`
	User                  interface{} `json:"user"`
}

func NewHandler(usecase *authusecase.Service) *Handler {
	return &Handler{usecase: usecase}
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

	result, err := h.usecase.Login(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, authdomain.ErrInvalidCredentials) {
			return apierror.Unauthorized("invalid username or password")
		}
		if errors.Is(err, authdomain.ErrAccountInactive) {
			return apierror.Unauthorized("account is inactive")
		}
		return apierror.Internal("failed to login")
	}

	return apiresponse.OK(c, toSessionResponse(result))
}

// Refresh godoc
// @Summary Exchange a refresh token for a new access/refresh token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param request body refreshRequest true "Refresh payload"
// @Success 200 {object} sessionResponse
// @Failure 400 {object} apierror.Error
// @Failure 401 {object} apierror.Error
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c fiber.Ctx) error {
	var req refreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	if req.RefreshToken == "" {
		return apierror.BadRequest("refresh_token is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	result, err := h.usecase.Refresh(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, authdomain.ErrRefreshTokenInvalid) {
			return apierror.Unauthorized("invalid or expired refresh token")
		}
		if errors.Is(err, authdomain.ErrAccountInactive) {
			return apierror.Unauthorized("account is inactive")
		}
		return apierror.Internal("failed to refresh session")
	}

	return apiresponse.OK(c, toSessionResponse(result))
}

// Logout godoc
// @Summary Revoke a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body refreshRequest true "Refresh payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apierror.Error
// @Router /auth/logout [post]
func (h *Handler) Logout(c fiber.Ctx) error {
	var req refreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return apierror.BadRequest("invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.usecase.Logout(ctx, req.RefreshToken); err != nil {
		return apierror.Internal("failed to logout")
	}

	return apiresponse.Message(c, "logged out")
}

func toSessionResponse(result authusecase.LoginResult) sessionResponse {
	return sessionResponse{
		AccessToken:           result.AccessToken,
		AccessTokenExpiresAt:  result.AccessTokenExpiresAt,
		RefreshToken:          result.RefreshToken,
		RefreshTokenExpiresAt: result.RefreshTokenExpiresAt,
		User:                  result.User,
	}
}
