package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/bookloop-alif/internal/domain/model"
)

type RefreshTokenRepo struct {
	db Querier
}

func NewRefreshTokenRepo(db Querier) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, rt *model.RefreshToken) (*model.RefreshToken, error) {
	const q = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, q, rt.UserID, rt.TokenHash, rt.ExpiresAt).
		Scan(&rt.ID, &rt.CreatedAt)
	if err != nil {
		return nil, mapPgError(err)
	}
	return rt, nil
}

func (r *RefreshTokenRepo) GetByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens WHERE token_hash = $1`

	rt := &model.RefreshToken{}
	err := r.db.QueryRow(ctx, q, hash).
		Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *RefreshTokenRepo) Revoke(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1`, id)
	return err
}
