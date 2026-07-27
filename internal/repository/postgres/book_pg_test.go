package postgres_test

import (
	"context"
	"testing"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/repository/postgres"
)

func TestBookRepo_CreateAndGet(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewBookRepo(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &model.Book{
		Title:  "Война и мир",
		Author: "Толстой",
		Genre:  "classic",
		Status: model.BookAvailable,
		Copies: 3,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero id")
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Title != "Война и мир" || got.Copies != 3 {
		t.Fatalf("unexpected book: %+v", got)
	}
}

func TestBookRepo_GetByID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewBookRepo(pool)

	_, err := repo.GetByID(context.Background(), 999999)
	if err != model.ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestBookRepo_DecrCopies(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewBookRepo(pool)
	ctx := context.Background()

	b, err := repo.Create(ctx, &model.Book{
		Title: "Test Book", Author: "A", Genre: "g",
		Status: model.BookAvailable, Copies: 1,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.DecrCopies(ctx, b.ID); err != nil {
		t.Fatalf("first decr should succeed: %v", err)
	}

	err = repo.DecrCopies(ctx, b.ID)
	if err != model.ErrNoSeats {
		t.Fatalf("want ErrNoSeats when copies=0, got %v", err)
	}

	got, err := repo.GetByID(ctx, b.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Copies != 0 {
		t.Fatalf("copies should be 0, got %d", got.Copies)
	}
}

func TestBookRepo_List_Filter(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewBookRepo(pool)
	ctx := context.Background()

	repo.Create(ctx, &model.Book{Title: "Go in Action", Author: "William", Genre: "tech", Status: model.BookAvailable, Copies: 1})
	repo.Create(ctx, &model.Book{Title: "Clean Code", Author: "Robert", Genre: "tech", Status: model.BookAvailable, Copies: 1})
	repo.Create(ctx, &model.Book{Title: "Dune", Author: "Frank", Genre: "scifi", Status: model.BookAvailable, Copies: 1})

	books, total, err := repo.List(ctx, repository.BookFilter{Genre: "tech", Limit: 20, Page: 1})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if total != 2 || len(books) != 2 {
		t.Fatalf("want 2 tech books, got total=%d len=%d", total, len(books))
	}
}
func TestBookRepo_Delete(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewBookRepo(pool)
	ctx := context.Background()

	b, _ := repo.Create(ctx, &model.Book{Title: "T", Author: "A", Genre: "g", Status: model.BookAvailable, Copies: 1})

	if err := repo.Delete(ctx, b.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := repo.GetByID(ctx, b.ID)
	if err != model.ErrNotFound {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
}

func TestBookRepo_Delete_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewBookRepo(pool)

	err := repo.Delete(context.Background(), 999999)
	if err != model.ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestBookRepo_IncrCopies(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewBookRepo(pool)
	ctx := context.Background()

	b, _ := repo.Create(ctx, &model.Book{Title: "T", Author: "A", Genre: "g", Status: model.BookAvailable, Copies: 1})

	if err := repo.IncrCopies(ctx, b.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := repo.GetByID(ctx, b.ID)
	if got.Copies != 2 {
		t.Fatalf("want copies=2, got %d", got.Copies)
	}
}

func TestBookRepo_Update_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewBookRepo(pool)

	err := repo.Update(context.Background(), &model.Book{ID: 999999, Title: "T", Author: "A", Genre: "g", Copies: 1})
	if err != model.ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
