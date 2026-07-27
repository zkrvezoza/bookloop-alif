package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/repository/postgres"
)

func seedUserAndBook(t *testing.T, pool postgres.Querier, copies int) (userID, bookID int64) {
	t.Helper()
	ctx := context.Background()

	userRepo := postgres.NewUserRepo(pool)
	bookRepo := postgres.NewBookRepo(pool)

	u, err := userRepo.Create(ctx, &model.User{
		Login:        "reader1",
		PasswordHash: "hash",
		Role:         model.RoleUser,
	})
	if err != nil || u == nil {
		t.Fatalf("seed user failed, err: %v, u: %v", err, u)
	}

	b, err := bookRepo.Create(ctx, &model.Book{
		Title: "Test Book", Author: "A", Genre: "g",
		Status: model.BookAvailable, Copies: copies,
	})
	if err != nil || b == nil {
		t.Fatalf("seed book failed, err: %v, b: %v", err, b)
	}

	return u.ID, b.ID
}

func TestLoanRepo_CreateAndGet(t *testing.T) {
	pool := setupTestDB(t)
	userID, bookID := seedUserAndBook(t, pool, 2)
	repo := postgres.NewLoanRepo(pool)
	ctx := context.Background()

	loan, err := repo.Create(ctx, &model.Loan{
		BookID: bookID,
		UserID: userID,
		DueAt:  time.Now().Add(14 * 24 * time.Hour),
		Status: model.LoanActive,
	})
	if err != nil || loan == nil {
		t.Fatalf("create failed, err: %v, loan: %v", err, loan)
	}
	if loan.ID == 0 {
		t.Fatal("expected non-zero id")
	}

	got, err := repo.GetByID(ctx, loan.ID)
	if err != nil || got == nil {
		t.Fatalf("get failed, err: %v, got: %v", err, got)
	}
	if got.Status != model.LoanActive || got.BookID != bookID {
		t.Fatalf("unexpected loan: %+v", got)
	}
}

func TestLoanRepo_MarkReturned(t *testing.T) {
	pool := setupTestDB(t)
	userID, bookID := seedUserAndBook(t, pool, 1)
	repo := postgres.NewLoanRepo(pool)
	ctx := context.Background()

	loan, err := repo.Create(ctx, &model.Loan{
		BookID: bookID, UserID: userID,
		DueAt: time.Now().Add(14 * 24 * time.Hour), Status: model.LoanActive,
	})
	if err != nil || loan == nil {
		t.Fatalf("create failed, err: %v, loan: %v", err, loan)
	}

	t.Run("marks active loan as returned", func(t *testing.T) {
		if err := repo.MarkReturned(ctx, loan.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := repo.GetByID(ctx, loan.ID)
		if err != nil || got == nil {
			t.Fatalf("get failed, err: %v, got: %v", err, got)
		}
		if got.Status != model.LoanReturned {
			t.Fatalf("want status returned, got %s", got.Status)
		}
		if got.ReturnedAt == nil {
			t.Fatal("expected returned_at to be set")
		}
	})

	t.Run("cannot return already-returned loan", func(t *testing.T) {
		err := repo.MarkReturned(ctx, loan.ID)
		if !errors.Is(err, model.ErrNotFound) {
			t.Fatalf("want ErrNotFound (already returned), got %v", err)
		}
	})
}

func TestLoanRepo_MarkOverdue(t *testing.T) {
	pool := setupTestDB(t)
	userID, bookID := seedUserAndBook(t, pool, 1)
	repo := postgres.NewLoanRepo(pool)
	ctx := context.Background()

	loan, err := repo.Create(ctx, &model.Loan{
		BookID: bookID, UserID: userID,
		DueAt: time.Now().Add(-24 * time.Hour), Status: model.LoanActive,
	})
	if err != nil || loan == nil {
		t.Fatalf("create failed, err: %v, loan: %v", err, loan)
	}

	if err := repo.MarkOverdue(ctx, loan.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetByID(ctx, loan.ID)
	if err != nil || got == nil {
		t.Fatalf("get failed, err: %v, got: %v", err, got)
	}
	if got.Status != model.LoanOverdue {
		t.Fatalf("want overdue, got %s", got.Status)
	}
}

func TestLoanRepo_CountActiveByBook(t *testing.T) {
	pool := setupTestDB(t)
	userID, bookID := seedUserAndBook(t, pool, 5)
	repo := postgres.NewLoanRepo(pool)
	ctx := context.Background()

	l1, err := repo.Create(ctx, &model.Loan{BookID: bookID, UserID: userID, DueAt: time.Now().Add(time.Hour), Status: model.LoanActive})
	if err != nil || l1 == nil {
		t.Fatalf("create l1 failed: %v", err)
	}
	_, _ = repo.Create(ctx, &model.Loan{BookID: bookID, UserID: userID, DueAt: time.Now().Add(time.Hour), Status: model.LoanActive})
	l3, err := repo.Create(ctx, &model.Loan{BookID: bookID, UserID: userID, DueAt: time.Now().Add(time.Hour), Status: model.LoanActive})
	if err != nil || l3 == nil {
		t.Fatalf("create l3 failed: %v", err)
	}
	_ = repo.MarkReturned(ctx, l3.ID)

	count, err := repo.CountActiveByBook(ctx, bookID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("want 2 active loans, got %d", count)
	}
}

func TestLoanRepo_List_FilterByUserAndStatus(t *testing.T) {
	pool := setupTestDB(t)
	userID, bookID := seedUserAndBook(t, pool, 5)
	repo := postgres.NewLoanRepo(pool)
	ctx := context.Background()

	l1, err := repo.Create(ctx, &model.Loan{BookID: bookID, UserID: userID, DueAt: time.Now().Add(time.Hour), Status: model.LoanActive})
	if err != nil || l1 == nil {
		t.Fatalf("create l1 failed: %v", err)
	}
	_, _ = repo.Create(ctx, &model.Loan{BookID: bookID, UserID: userID, DueAt: time.Now().Add(time.Hour), Status: model.LoanActive})
	_ = repo.MarkReturned(ctx, l1.ID)

	loans, total, err := repo.List(ctx, repository.LoanFilter{UserID: userID, Status: string(model.LoanActive), Limit: 20, Page: 1})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if total != 1 || len(loans) != 1 {
		t.Fatalf("want 1 active loan, got total=%d len=%d", total, len(loans))
	}
}

func TestLoanRepo_ExtendDueDate(t *testing.T) {
	pool := setupTestDB(t)
	userID, bookID := seedUserAndBook(t, pool, 1)
	repo := postgres.NewLoanRepo(pool)
	ctx := context.Background()

	loan, err := repo.Create(ctx, &model.Loan{
		BookID: bookID, UserID: userID,
		DueAt: time.Now().Add(24 * time.Hour), Status: model.LoanActive,
	})
	if err != nil || loan == nil {
		t.Fatalf("create failed, err: %v, loan: %v", err, loan)
	}

	newDue := time.Now().Add(48 * time.Hour).Format(time.RFC3339)
	if err := repo.ExtendDueDate(ctx, loan.ID, newDue); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetByID(ctx, loan.ID)
	if err != nil || got == nil {
		t.Fatalf("get failed, err: %v, got: %v", err, got)
	}
	if got.DueAt.Before(loan.DueAt) {
		t.Fatalf("due date should be extended, got %v", got.DueAt)
	}
}

func TestLoanRepo_ExtendDueDate_NotFoundWhenNotActive(t *testing.T) {
	pool := setupTestDB(t)
	userID, bookID := seedUserAndBook(t, pool, 1)
	repo := postgres.NewLoanRepo(pool)
	ctx := context.Background()

	loan, err := repo.Create(ctx, &model.Loan{BookID: bookID, UserID: userID, DueAt: time.Now(), Status: model.LoanActive})
	if err != nil || loan == nil {
		t.Fatalf("create failed, err: %v, loan: %v", err, loan)
	}
	_ = repo.MarkReturned(ctx, loan.ID)

	err = repo.ExtendDueDate(ctx, loan.ID, time.Now().Format(time.RFC3339))
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
