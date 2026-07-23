package repository

import (
	"context"

	"github.com/bookloop-alif/internal/domain/model"
)

type UserRepo interface {
	Create(ctx context.Context, u *model.User) (*model.User, error)
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
}
