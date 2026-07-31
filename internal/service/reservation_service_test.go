package service

import (
	"context"
	"errors"
	"testing"

	"github.com/bookloop-alif/internal/domain/model"
)

type mockReservationRepo struct {
	createFn        func(context.Context, *model.Reservation) (*model.Reservation, error)
	listByUserFn    func(context.Context, int64) ([]model.Reservation, error)
	cancelFn        func(context.Context, int64) error
	nextWaitingFn   func(context.Context, int64) (*model.Reservation, error)
	markFulfilledFn func(context.Context, int64) error
}

func (m *mockReservationRepo) Create(
	ctx context.Context,
	reservation *model.Reservation,
) (*model.Reservation, error) {
	if m.createFn == nil {
		panic("unexpected call to Create")
	}

	return m.createFn(ctx, reservation)
}

func (m *mockReservationRepo) ListByUser(
	ctx context.Context,
	userID int64,
) ([]model.Reservation, error) {
	if m.listByUserFn == nil {
		panic("unexpected call to ListByUser")
	}

	return m.listByUserFn(ctx, userID)
}

func (m *mockReservationRepo) Cancel(
	ctx context.Context,
	reservationID int64,
) error {
	if m.cancelFn == nil {
		panic("unexpected call to Cancel")
	}

	return m.cancelFn(ctx, reservationID)
}

func (m *mockReservationRepo) NextWaiting(
	ctx context.Context,
	bookID int64,
) (*model.Reservation, error) {
	if m.nextWaitingFn == nil {
		panic("unexpected call to NextWaiting")
	}

	return m.nextWaitingFn(ctx, bookID)
}

func (m *mockReservationRepo) MarkFulfilled(
	ctx context.Context,
	reservationID int64,
) error {
	if m.markFulfilledFn == nil {
		panic("unexpected call to MarkFulfilled")
	}

	return m.markFulfilledFn(ctx, reservationID)
}

func TestReservationService_Reserve_OK(t *testing.T) {
	bookRepo := &mockBookRepo{
		getByIDFn: func(
			ctx context.Context,
			bookID int64,
		) (*model.Book, error) {
			return &model.Book{
				ID:     bookID,
				Copies: 0,
			}, nil
		},
	}

	resRepo := &mockReservationRepo{
		createFn: func(
			ctx context.Context,
			reservation *model.Reservation,
		) (*model.Reservation, error) {
			reservation.ID = 10
			return reservation, nil
		},
	}

	svc := NewReservationService(resRepo, bookRepo)

	got, err := svc.Reserve(context.Background(), 3, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != 10 {
		t.Fatalf("expected reservation ID 10, got %d", got.ID)
	}

	if got.UserID != 3 {
		t.Fatalf("expected user ID 3, got %d", got.UserID)
	}

	if got.BookID != 7 {
		t.Fatalf("expected book ID 7, got %d", got.BookID)
	}

	if got.Status != model.ReservationWaiting {
		t.Fatalf(
			"expected status %q, got %q",
			model.ReservationWaiting,
			got.Status,
		)
	}
}

func TestReservationService_Reserve_BookRepoError(t *testing.T) {
	expectedErr := errors.New("book database failed")

	bookRepo := &mockBookRepo{
		getByIDFn: func(
			ctx context.Context,
			bookID int64,
		) (*model.Book, error) {
			return nil, expectedErr
		},
	}

	svc := NewReservationService(
		&mockReservationRepo{},
		bookRepo,
	)

	got, err := svc.Reserve(context.Background(), 3, 7)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatalf("expected nil reservation, got %#v", got)
	}
}

func TestReservationService_Reserve_BookAvailable(t *testing.T) {
	bookRepo := &mockBookRepo{
		getByIDFn: func(
			ctx context.Context,
			bookID int64,
		) (*model.Book, error) {
			return &model.Book{
				ID:     bookID,
				Copies: 2,
			}, nil
		},
	}

	svc := NewReservationService(
		&mockReservationRepo{},
		bookRepo,
	)

	got, err := svc.Reserve(context.Background(), 3, 7)

	if !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}

	if got != nil {
		t.Fatalf("expected nil reservation, got %#v", got)
	}
}

func TestReservationService_MyReservations_OK(t *testing.T) {
	expected := []model.Reservation{
		{
			ID:     1,
			UserID: 3,
			BookID: 7,
			Status: model.ReservationWaiting,
		},
	}

	resRepo := &mockReservationRepo{
		listByUserFn: func(
			ctx context.Context,
			userID int64,
		) ([]model.Reservation, error) {
			if userID != 3 {
				t.Fatalf("expected user ID 3, got %d", userID)
			}

			return expected, nil
		},
	}

	svc := NewReservationService(resRepo, nil)

	got, err := svc.MyReservations(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected one reservation, got %d", len(got))
	}
}

func TestReservationService_Cancel_OK(t *testing.T) {
	var cancelledID int64

	resRepo := &mockReservationRepo{
		cancelFn: func(
			ctx context.Context,
			reservationID int64,
		) error {
			cancelledID = reservationID
			return nil
		},
	}

	svc := NewReservationService(resRepo, nil)

	err := svc.Cancel(context.Background(), 3, 15)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cancelledID != 15 {
		t.Fatalf("expected reservation ID 15, got %d", cancelledID)
	}
}

func TestReservationService_Cancel_RepoError(t *testing.T) {
	expectedErr := errors.New("cancel failed")

	resRepo := &mockReservationRepo{
		cancelFn: func(
			ctx context.Context,
			reservationID int64,
		) error {
			return expectedErr
		},
	}

	svc := NewReservationService(resRepo, nil)

	err := svc.Cancel(context.Background(), 3, 15)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
