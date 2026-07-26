package repository

import (
	"context"

	"github.com/bookloop-alif/internal/domain/model"
)

type RefreshTokenRepo interface {
	Create(ctx context.Context, rt *model.RefreshToken) (*model.RefreshToken, error)
	GetByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	Revoke(ctx context.Context, id int64) error
}
