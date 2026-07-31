package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/repository/postgres"
)

func TestUnitOfWork_WithinTx_CommitsOnSuccess(t *testing.T) {
	pool := setupTestDB(t)

	// 1. Создаём TxManager из pool
	txManager := postgres.NewTxManager(pool)
	// 2. Передаём txManager в NewUnitOfWork
	uow := postgres.NewUnitOfWork(txManager)

	ctx := context.Background()

	var bookID int64
	err := uow.WithinTx(ctx, func(books repository.BookRepo, loans repository.LoanRepo) error {
		b, err := books.Create(ctx, &model.Book{Title: "T", Author: "A", Genre: "g", Status: model.BookAvailable, Copies: 1})
		if err != nil {
			return err
		}
		bookID = b.ID
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repo := postgres.NewBookRepo(pool)
	got, err := repo.GetByID(ctx, bookID)
	if err != nil {
		t.Fatalf("book should exist after commit: %v", err)
	}
	if got.Title != "T" {
		t.Fatalf("unexpected book: %+v", got)
	}
}

func TestUnitOfWork_WithinTx_RollsBackOnError(t *testing.T) {
	pool := setupTestDB(t)
	txManager := postgres.NewTxManager(pool)
	uow := postgres.NewUnitOfWork(txManager)

	ctx := context.Background()

	wantErr := errors.New("intentional failure")
	var bookID int64

	err := uow.WithinTx(ctx, func(books repository.BookRepo, loans repository.LoanRepo) error {
		b, _ := books.Create(ctx, &model.Book{Title: "T", Author: "A", Genre: "g", Status: model.BookAvailable, Copies: 1})
		if b != nil {
			bookID = b.ID
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("want wrapped intentional error, got %v", err)
	}

	repo := postgres.NewBookRepo(pool)
	_, err = repo.GetByID(ctx, bookID)
	if err != model.ErrNotFound {
		t.Fatalf("book should NOT exist after rollback, got %v", err)
	}
}
