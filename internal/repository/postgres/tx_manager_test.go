package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/repository/postgres"
)

func TestTxManager_WithTx_CommitsOnSuccess(t *testing.T) {
	pool := setupTestDB(t)
	tx := postgres.NewTxManager(pool)
	ctx := context.Background()

	var bookID int64
	err := tx.WithTx(ctx, func(q postgres.Querier) error {
		repo := postgres.NewBookRepo(q)
		b, err := repo.Create(ctx, &model.Book{Title: "T", Author: "A", Genre: "g", Status: model.BookAvailable, Copies: 1})
		bookID = b.ID
		return err
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

func TestTxManager_WithTx_RollsBackOnError(t *testing.T) {
	pool := setupTestDB(t)
	tx := postgres.NewTxManager(pool)
	ctx := context.Background()

	wantErr := errors.New("intentional failure")
	var bookID int64

	err := tx.WithTx(ctx, func(q postgres.Querier) error {
		repo := postgres.NewBookRepo(q)
		b, _ := repo.Create(ctx, &model.Book{Title: "T", Author: "A", Genre: "g", Status: model.BookAvailable, Copies: 1})
		bookID = b.ID
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
