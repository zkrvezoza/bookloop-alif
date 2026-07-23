package postgres

import (
	"context"
	"errors"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/jackc/pgx/v5"
)

type ReservationRepo struct {
	db Querier
}

func NewReservationRepo(db Querier) *ReservationRepo {
	return &ReservationRepo{db: db}
}

func (r *ReservationRepo) Create(ctx context.Context, res *model.Reservation) (*model.Reservation, error) {
	const q = `
		INSERT INTO reservations (book_id, user_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, q, res.BookID, res.UserID, res.Status).
		Scan(&res.ID, &res.CreatedAt)
	if err != nil {
		return nil, mapPgError(err)
	}
	return res, nil
}

func (r *ReservationRepo) NextWaiting(ctx context.Context, bookID int64) (*model.Reservation, error) {
	const q = `
		SELECT id, book_id, user_id, created_at, status
		FROM reservations
		WHERE book_id = $1 AND status = 'waiting'
		ORDER BY created_at
		LIMIT 1
		FOR UPDATE SKIP LOCKED`

	res := &model.Reservation{}
	err := r.db.QueryRow(ctx, q, bookID).
		Scan(&res.ID, &res.BookID, &res.UserID, &res.CreatedAt, &res.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *ReservationRepo) MarkFulfilled(ctx context.Context, id int64) error {
	const q = `UPDATE reservations SET status = 'fulfilled' WHERE id = $1 AND status = 'waiting'`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *ReservationRepo) Cancel(ctx context.Context, id int64) error {
	const q = `UPDATE reservations SET status = 'cancelled' WHERE id = $1 AND status = 'waiting'`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *ReservationRepo) ListByUser(ctx context.Context, userID int64) ([]model.Reservation, error) {
	const q = `
		SELECT id, book_id, user_id, created_at, status
		FROM reservations WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Reservation
	for rows.Next() {
		var res model.Reservation
		if err := rows.Scan(&res.ID, &res.BookID, &res.UserID, &res.CreatedAt, &res.Status); err != nil {
			return nil, err
		}
		list = append(list, res)
	}
	return list, rows.Err()
}
