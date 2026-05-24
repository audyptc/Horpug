package domain

import "context"

type RefreshTokenRepository interface {
	Save(ctx context.Context, token *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, hash string) error
}
