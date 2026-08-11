package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	authdomain "apihorpug/internal/features/auth/domain"
	userdomain "apihorpug/internal/features/user/domain"
	platformjwt "apihorpug/internal/platform/jwt"

	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	FindByLogin(ctx context.Context, login string) (userdomain.User, error)
}

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	User      userdomain.User
}

type Service struct {
	repo      Repository
	jwtSecret string
	jwtTTL    time.Duration
}

func New(repo Repository, jwtSecret string, jwtTTL time.Duration) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret, jwtTTL: jwtTTL}
}

func (s *Service) Login(ctx context.Context, login, password string) (LoginResult, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return LoginResult{}, authdomain.ErrInvalidCredentials
	}

	user, err := s.repo.FindByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return LoginResult{}, authdomain.ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if !user.IsActive {
		return LoginResult{}, authdomain.ErrAccountInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return LoginResult{}, authdomain.ErrInvalidCredentials
	}

	token, expiresAt, err := platformjwt.Generate(s.jwtSecret, s.jwtTTL, user.ID, user.RoleID, user.Username)
	if err != nil {
		return LoginResult{}, err
	}

	user.Password = ""
	return LoginResult{Token: token, ExpiresAt: expiresAt, User: user}, nil
}
