package postgres

import (
	"context"
	"errors"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	db Querier
}

func NewUserRepo(db Querier) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *model.User) (*model.User, error) {
	const q = `
		INSERT INTO users (login, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, q, u.Login, u.PasswordHash, u.Role).
		Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return nil, mapPgError(err)
	}
	return u, nil
}

func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	const q = `
		SELECT id, login, password_hash, role, created_at
		FROM users WHERE login = $1`

	u := &model.User{}
	err := r.db.QueryRow(ctx, q, login).
		Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	const q = `
		SELECT id, login, password_hash, role, created_at
		FROM users WHERE id = $1`

	u := &model.User{}
	err := r.db.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}
