package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	authdomain "apihorpug/internal/features/auth/domain"
	userdomain "apihorpug/internal/features/user/domain"
	platformjwt "apihorpug/internal/platform/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	FindByLogin(ctx context.Context, login string) (userdomain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (userdomain.User, error)
}

type TokenRepository interface {
	SaveRefreshToken(ctx context.Context, token authdomain.RefreshToken) error
	FindRefreshToken(ctx context.Context, tokenHash string) (authdomain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
}

// ActivityLogger records login/logout events for the audit trail. Failures to
// record are logged but never block the auth flow itself.
type ActivityLogger interface {
	Create(ctx context.Context, input activitylogusecase.CreateInput) (activitylogdomain.ActivityLog, error)
}

type LoginResult struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	User                  userdomain.User
}

type Service struct {
	userRepo    UserRepository
	tokenRepo   TokenRepository
	activityLog ActivityLogger
	jwtSecret   string
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func New(userRepo UserRepository, tokenRepo TokenRepository, activityLog ActivityLogger, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		activityLog: activityLog,
		jwtSecret:   jwtSecret,
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
	}
}

func (s *Service) Login(ctx context.Context, login, password, ipAddress string) (LoginResult, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return LoginResult{}, authdomain.ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByLogin(ctx, login)
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

	result, err := s.issueSession(ctx, user)
	if err != nil {
		return LoginResult{}, err
	}

	s.recordActivity(ctx, &user.ID, "LOGIN", "User logged in", ipAddress)
	return result, nil
}

// Refresh rotates a refresh token: the presented token is revoked and a new
// access/refresh pair is issued, so a stolen-but-already-used token stops working.
func (s *Service) Refresh(ctx context.Context, rawToken string) (LoginResult, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return LoginResult{}, authdomain.ErrRefreshTokenInvalid
	}

	stored, err := s.tokenRepo.FindRefreshToken(ctx, hashToken(rawToken))
	if err != nil {
		return LoginResult{}, err
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) {
		return LoginResult{}, authdomain.ErrRefreshTokenInvalid
	}

	user, err := s.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return LoginResult{}, authdomain.ErrRefreshTokenInvalid
		}
		return LoginResult{}, err
	}
	if !user.IsActive {
		return LoginResult{}, authdomain.ErrAccountInactive
	}

	if err := s.tokenRepo.RevokeRefreshToken(ctx, stored.ID); err != nil {
		return LoginResult{}, err
	}

	return s.issueSession(ctx, user)
}

// Logout revokes a refresh token. Already-invalid tokens are treated as
// success since the end state (no usable session) is the same.
func (s *Service) Logout(ctx context.Context, rawToken, ipAddress string) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil
	}

	stored, err := s.tokenRepo.FindRefreshToken(ctx, hashToken(rawToken))
	if err != nil {
		if errors.Is(err, authdomain.ErrRefreshTokenInvalid) {
			return nil
		}
		return err
	}

	if err := s.tokenRepo.RevokeRefreshToken(ctx, stored.ID); err != nil {
		return err
	}

	s.recordActivity(ctx, &stored.UserID, "LOGOUT", "User logged out", ipAddress)
	return nil
}

// recordActivity is best-effort: a failure to write the audit trail must
// never fail the login/logout flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}

	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "auth",
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func (s *Service) issueSession(ctx context.Context, user userdomain.User) (LoginResult, error) {
	accessToken, accessExpiresAt, err := platformjwt.Generate(s.jwtSecret, s.accessTTL, user.ID, user.RoleID, user.Username)
	if err != nil {
		return LoginResult{}, err
	}

	rawRefreshToken, err := generateRandomToken()
	if err != nil {
		return LoginResult{}, err
	}
	refreshExpiresAt := time.Now().Add(s.refreshTTL)

	if err := s.tokenRepo.SaveRefreshToken(ctx, authdomain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hashToken(rawRefreshToken),
		ExpiresAt: refreshExpiresAt,
	}); err != nil {
		return LoginResult{}, err
	}

	user.Password = ""
	return LoginResult{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          rawRefreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
		User:                  user,
	}, nil
}

func generateRandomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
