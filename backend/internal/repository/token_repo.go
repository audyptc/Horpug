package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type TokenRepo struct {
	db *database.DB
}

func NewTokenRepo(db *database.DB) *TokenRepo {
	return &TokenRepo{db: db}
}

func (r *TokenRepo) Save(ctx context.Context, token *domain.RefreshToken) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)`,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt)
	return err
}

func (r *TokenRepo) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	t := &domain.RefreshToken{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens WHERE token_hash = $1`, hash).
		Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("token not found")
	}
	return t, err
}

func (r *TokenRepo) Revoke(ctx context.Context, hash string) error {
	result, err := r.db.Pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL`, hash)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("token not found or already revoked")
	}
	return nil
}

func (r *TokenRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}
