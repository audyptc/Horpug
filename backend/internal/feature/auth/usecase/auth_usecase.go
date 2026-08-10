package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/auth/domain"
	userdomain "apigofiberhorpug/internal/feature/user/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultAccessTokenDuration  = 15 * time.Minute
	defaultRefreshTokenDuration = 7 * 24 * time.Hour
)

type Claims struct {
	UserID               string              `json:"user_id"`
	Email                string              `json:"email"`
	RoleName             string              `json:"role_name"`
	Permissions          []string            `json:"permissions"`
	GlobalPermissions    []string            `json:"global_permissions"`
	DormitoryPermissions map[string][]string `json:"dormitory_permissions"`
	jwt.RegisteredClaims
}

type AuthUseCase struct {
	userRepo             userdomain.UserRepository
	tokenRepo            domain.RefreshTokenRepository
	secretKey            []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewAuthUseCase(
	userRepo userdomain.UserRepository,
	tokenRepo domain.RefreshTokenRepository,
	secretKey string,
	accessTokenDuration, refreshTokenDuration time.Duration,
) *AuthUseCase {
	if accessTokenDuration <= 0 {
		accessTokenDuration = defaultAccessTokenDuration
	}
	if refreshTokenDuration <= 0 {
		refreshTokenDuration = defaultRefreshTokenDuration
	}

	return &AuthUseCase{
		userRepo:             userRepo,
		tokenRepo:            tokenRepo,
		secretKey:            []byte(secretKey),
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (uc *AuthUseCase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apierror.Unauthorized("invalid credentials")
	}
	if !user.IsActive {
		return nil, apierror.Forbidden("account is disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apierror.Unauthorized("invalid credentials")
	}

	globalPermissions, err := uc.userRepo.GetPermissions(ctx, user.ID)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	dormitoryPermissions, err := uc.userRepo.GetDormitoryPermissions(ctx, user.ID)
	if err != nil {
		return nil, apierror.Internal(err)
	}

	role, err := uc.userRepo.GetRole(ctx, user.ID)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	roleName := ""
	if role != nil {
		roleName = role.Name
	}
	mergedPermissions := mergePermissions(globalPermissions, dormitoryPermissions)
	if roleName == "" && len(dormitoryPermissions) == 0 {
		return nil, apierror.Forbidden("account has no active role")
	}

	accessToken, err := uc.generateAccessToken(user, roleName, mergedPermissions, globalPermissions, dormitoryPermissions)
	if err != nil {
		return nil, apierror.Internal(err)
	}

	rawRefresh := uuid.New().String()
	rt := &domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: hashString(rawRefresh),
		ExpiresAt: time.Now().Add(uc.refreshTokenDuration),
	}
	if err := uc.tokenRepo.Save(ctx, rt); err != nil {
		return nil, apierror.Internal(err)
	}

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(uc.accessTokenDuration.Seconds()),
	}, nil
}

func (uc *AuthUseCase) Refresh(ctx context.Context, rawToken string) (*domain.LoginResponse, error) {
	rt, err := uc.tokenRepo.FindByHash(ctx, hashString(rawToken))
	if err != nil {
		return nil, apierror.Unauthorized("invalid refresh token")
	}
	if rt.RevokedAt != nil {
		// Fix 5: token ที่ถูก revoke แล้วถูกนำมาใช้ซ้ำ หมายความว่า token อาจถูกขโมย
		// revoke ทุก session ของ user นี้ทันที
		_ = uc.tokenRepo.RevokeAllForUser(ctx, rt.UserID)
		return nil, apierror.Unauthorized("refresh token has been revoked")
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, apierror.Unauthorized("refresh token has expired")
	}

	if err := uc.tokenRepo.Revoke(ctx, hashString(rawToken)); err != nil {
		return nil, apierror.Internal(err)
	}

	user, err := uc.userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	if !user.IsActive {
		return nil, apierror.Forbidden("account is disabled")
	}

	globalPermissions, err := uc.userRepo.GetPermissions(ctx, user.ID)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	dormitoryPermissions, err := uc.userRepo.GetDormitoryPermissions(ctx, user.ID)
	if err != nil {
		return nil, apierror.Internal(err)
	}

	refreshRole, err := uc.userRepo.GetRole(ctx, user.ID)
	if err != nil {
		return nil, apierror.Internal(err)
	}
	refreshRoleName := ""
	if refreshRole != nil {
		refreshRoleName = refreshRole.Name
	}
	mergedPermissions := mergePermissions(globalPermissions, dormitoryPermissions)
	if refreshRoleName == "" && len(dormitoryPermissions) == 0 {
		return nil, apierror.Forbidden("account has no active role")
	}

	accessToken, err := uc.generateAccessToken(user, refreshRoleName, mergedPermissions, globalPermissions, dormitoryPermissions)
	if err != nil {
		return nil, apierror.Internal(err)
	}

	rawRefresh := uuid.New().String()
	newRT := &domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: hashString(rawRefresh),
		ExpiresAt: time.Now().Add(uc.refreshTokenDuration),
	}
	if err := uc.tokenRepo.Save(ctx, newRT); err != nil {
		return nil, apierror.Internal(err)
	}

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(uc.accessTokenDuration.Seconds()),
	}, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	_ = uc.tokenRepo.Revoke(ctx, hashString(rawToken))
	return nil
}

func (uc *AuthUseCase) RefreshMaxAge() int {
	return int(uc.refreshTokenDuration.Seconds())
}

func (uc *AuthUseCase) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apierror.Unauthorized("unexpected signing method")
		}
		return uc.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apierror.Unauthorized("invalid token")
	}
	return claims, nil
}

func (uc *AuthUseCase) generateAccessToken(user *userdomain.User, roleName string, permissions, globalPermissions []string, dormitoryPermissions map[string][]string) (string, error) {
	claims := &Claims{
		UserID:               user.ID,
		Email:                user.Email,
		RoleName:             roleName,
		Permissions:          permissions,
		GlobalPermissions:    globalPermissions,
		DormitoryPermissions: dormitoryPermissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(uc.secretKey)
}

func mergePermissions(globalPermissions []string, dormitoryPermissions map[string][]string) []string {
	seen := make(map[string]struct{}, len(globalPermissions))
	merged := make([]string, 0, len(globalPermissions))

	appendUnique := func(values []string) {
		for _, value := range values {
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			merged = append(merged, value)
		}
	}

	appendUnique(globalPermissions)
	for _, permissions := range dormitoryPermissions {
		appendUnique(permissions)
	}

	return merged
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
