package postgres_test

import (
	"context"
	"testing"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/repository/postgres"
)

func TestStatsRepo_LibraryStats(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepo(pool)
	bookRepo := postgres.NewBookRepo(pool)
	loanRepo := postgres.NewLoanRepo(pool)
	resRepo := postgres.NewReservationRepo(pool)
	statsRepo := postgres.NewStatsRepo(pool)

	u, _ := userRepo.Create(ctx, &model.User{Login: "reader1", PasswordHash: "h", Role: model.RoleUser})
	b, _ := bookRepo.Create(ctx, &model.Book{Title: "T", Author: "A", Genre: "g", Status: model.BookAvailable, Copies: 2})

	_, _ = loanRepo.Create(ctx, &model.Loan{BookID: b.ID, UserID: u.ID, Status: model.LoanActive})
	overdue, _ := loanRepo.Create(ctx, &model.Loan{BookID: b.ID, UserID: u.ID, Status: model.LoanActive})
	_ = loanRepo.MarkOverdue(ctx, overdue.ID)
	_, _ = resRepo.Create(ctx, &model.Reservation{BookID: b.ID, UserID: u.ID, Status: model.ReservationWaiting})

	stats, err := statsRepo.LibraryStats(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.BooksTotal != 1 {
		t.Fatalf("want 1 book, got %d", stats.BooksTotal)
	}
	if stats.LoansActive != 1 {
		t.Fatalf("want 1 active loan, got %d", stats.LoansActive)
	}
	if stats.LoansOverdue != 1 {
		t.Fatalf("want 1 overdue loan, got %d", stats.LoansOverdue)
	}
	if stats.ReservationsWait != 1 {
		t.Fatalf("want 1 waiting reservation, got %d", stats.ReservationsWait)
	}
}
