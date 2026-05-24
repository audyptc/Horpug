package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"apigofiberhorpug/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
)

type Claims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

type AuthUseCase struct {
	userRepo  domain.UserRepository
	tokenRepo domain.RefreshTokenRepository
	secretKey []byte
}

func NewAuthUseCase(userRepo domain.UserRepository, tokenRepo domain.RefreshTokenRepository, secretKey string) *AuthUseCase {
	return &AuthUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		secretKey: []byte(secretKey),
	}
}

func (uc *AuthUseCase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	permissions, err := uc.userRepo.GetPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	accessToken, err := uc.generateAccessToken(user, permissions)
	if err != nil {
		return nil, err
	}

	rawRefresh := uuid.New().String()
	rt := &domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: hashString(rawRefresh),
		ExpiresAt: time.Now().Add(refreshTokenDuration),
	}
	if err := uc.tokenRepo.Save(ctx, rt); err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(accessTokenDuration.Seconds()),
	}, nil
}

func (uc *AuthUseCase) Refresh(ctx context.Context, rawToken string) (*domain.LoginResponse, error) {
	if rawToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	rt, err := uc.tokenRepo.FindByHash(ctx, hashString(rawToken))
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}
	if rt.RevokedAt != nil {
		return nil, fmt.Errorf("refresh token has been revoked")
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	// rotate: revoke old, issue new
	if err := uc.tokenRepo.Revoke(ctx, hashString(rawToken)); err != nil {
		return nil, err
	}

	user, err := uc.userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	permissions, err := uc.userRepo.GetPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	accessToken, err := uc.generateAccessToken(user, permissions)
	if err != nil {
		return nil, err
	}

	rawRefresh := uuid.New().String()
	newRT := &domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: hashString(rawRefresh),
		ExpiresAt: time.Now().Add(refreshTokenDuration),
	}
	if err := uc.tokenRepo.Save(ctx, newRT); err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(accessTokenDuration.Seconds()),
	}, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	// silently ignore "already revoked" errors on logout
	_ = uc.tokenRepo.Revoke(ctx, hashString(rawToken))
	return nil
}

func (uc *AuthUseCase) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return uc.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (uc *AuthUseCase) generateAccessToken(user *domain.User, permissions []string) (string, error) {
	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(uc.secretKey)
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
