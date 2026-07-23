package repository

import (
	"context"

	"github.com/bookloop-alif/internal/domain/model"
)

type ReservationRepo interface {
	Create(ctx context.Context, r *model.Reservation) (*model.Reservation, error)
	NextWaiting(ctx context.Context, bookID int64) (*model.Reservation, error)
	MarkFulfilled(ctx context.Context, id int64) error
	Cancel(ctx context.Context, id int64) error
	ListByUser(ctx context.Context, userID int64) ([]model.Reservation, error)
}
