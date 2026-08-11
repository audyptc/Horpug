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

type loginResponse struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      interface{} `json:"user"`
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
// @Success 200 {object} loginResponse
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

	return apiresponse.OK(c, loginResponse{
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt,
		User:      result.User,
	})
}
