package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/repository/postgres"
)

func TestReservationRepo_CreateAndListByUser(t *testing.T) {
	pool := setupTestDB(t)
	userID, bookID := seedUserAndBook(t, pool, 0)
	repo := postgres.NewReservationRepo(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &model.Reservation{
		BookID: bookID, UserID: userID, Status: model.ReservationWaiting,
	})
	if err != nil || created == nil {
		t.Fatalf("create failed, err: %v, created: %v", err, created)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero id")
	}

	list, err := repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func TestReservationRepo_NextWaiting_FIFOOrder(t *testing.T) {
	pool := setupTestDB(t)
	userID, bookID := seedUserAndBook(t, pool, 0)
	repo := postgres.NewReservationRepo(pool)
	ctx := context.Background()

	userRepo := postgres.NewUserRepo(pool)
	u2, err := userRepo.Create(ctx, &model.User{Login: "reader2", PasswordHash: "hash", Role: model.RoleUser})
	if err != nil || u2 == nil {
		t.Fatalf("create second user failed, err: %v, u2: %v", err, u2)
	}

	first, err := repo.Create(ctx, &model.Reservation{BookID: bookID, UserID: userID, Status: model.ReservationWaiting})
	if err != nil || first == nil {
		t.Fatalf("create first reservation failed, err: %v, first: %v", err, first)
	}

	time.Sleep(10 * time.Millisecond)

	_, err = repo.Create(ctx, &model.Reservation{BookID: bookID, UserID: u2.ID, Status: model.ReservationWaiting})
	if err != nil {
		t.Fatalf("create second reservation: %v", err)
	}

	next, err := repo.NextWaiting(ctx, bookID)
	if err != nil || next == nil {
		t.Fatalf("NextWaiting failed, err: %v, next: %v", err, next)
	}
	if next.ID != first.ID {
		t.Fatalf("want first-created reservation (FIFO), got id=%d want id=%d", next.ID, first.ID)
	}
}

func TestReservationRepo_NextWaiting_NoneWaiting(t *testing.T) {
	pool := setupTestDB(t)
	_, bookID := seedUserAndBook(t, pool, 0)
	repo := postgres.NewReservationRepo(pool)

	_, err := repo.NextWaiting(context.Background(), bookID)
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("want ErrNotFound when no reservations, got %v", err)
	}
}

func TestReservationRepo_MarkFulfilledAndCancel(t *testing.T) {
	pool := setupTestDB(t)
	userID, bookID := seedUserAndBook(t, pool, 0)
	repo := postgres.NewReservationRepo(pool)
	ctx := context.Background()

	r, err := repo.Create(ctx, &model.Reservation{BookID: bookID, UserID: userID, Status: model.ReservationWaiting})
	if err != nil || r == nil {
		t.Fatalf("create failed, err: %v, r: %v", err, r)
	}

	if err := repo.MarkFulfilled(ctx, r.ID); err != nil {
		t.Fatalf("mark fulfilled failed: %v", err)
	}

	if err := repo.MarkFulfilled(ctx, r.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("want ErrNotFound on already-fulfilled, got %v", err)
	}
}

func TestReservationRepo_Cancel(t *testing.T) {
	pool := setupTestDB(t)
	userID, bookID := seedUserAndBook(t, pool, 0)
	repo := postgres.NewReservationRepo(pool)
	ctx := context.Background()

	r, err := repo.Create(ctx, &model.Reservation{BookID: bookID, UserID: userID, Status: model.ReservationWaiting})
	if err != nil || r == nil {
		t.Fatalf("create failed, err: %v, r: %v", err, r)
	}

	if err := repo.Cancel(ctx, r.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = repo.Cancel(ctx, r.ID)
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("want ErrNotFound on already-cancelled, got %v", err)
	}
}
