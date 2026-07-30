package service

import (
	"context"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
)

type ReservationService struct {
	res   repository.ReservationRepo
	books repository.BookRepo
}

func NewReservationService(res repository.ReservationRepo, books repository.BookRepo) *ReservationService {
	return &ReservationService{res: res, books: books}
}

func (s *ReservationService) Reserve(ctx context.Context, userID, bookID int64) (*model.Reservation, error) {
	b, err := s.books.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}
	if b.Copies > 0 {
		return nil, model.ErrInvalid
	}

	return s.res.Create(ctx, &model.Reservation{
		BookID: bookID,
		UserID: userID,
		Status: model.ReservationWaiting,
	})
}

func (s *ReservationService) MyReservations(ctx context.Context, userID int64) ([]model.Reservation, error) {
	return s.res.ListByUser(ctx, userID)
}

func (s *ReservationService) Cancel(ctx context.Context, userID, reservationID int64) error {
	return s.res.Cancel(ctx, reservationID)
}
