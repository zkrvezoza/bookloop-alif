package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/service"
)

func TestBookService_Create(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		copies  int
		wantErr error
	}{
		{"valid", "Dune", 3, nil},
		{"empty title", "", 3, model.ErrInvalid},
		{"zero copies", "Dune", 0, model.ErrInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewBookService(newFakeBookRepo(), newFakeLoanRepo())
			_, err := svc.Create(context.Background(), tt.title, "Author", "genre", tt.copies)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestBookService_Get(t *testing.T) {
	books, loans := newFakeBookRepo(), newFakeLoanRepo()
	svc := service.NewBookService(books, loans)
	ctx := context.Background()

	b, _ := svc.Create(ctx, "Dune", "H", "scifi", 1)

	got, err := svc.Get(ctx, b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "Dune" {
		t.Fatalf("unexpected book: %+v", got)
	}
}

func TestBookService_Get_NotFound(t *testing.T) {
	svc := service.NewBookService(newFakeBookRepo(), newFakeLoanRepo())
	_, err := svc.Get(context.Background(), 999)
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestBookService_Update_RejectsCopiesBelowActiveLoans(t *testing.T) {
	books, loans := newFakeBookRepo(), newFakeLoanRepo()
	svc := service.NewBookService(books, loans)
	ctx := context.Background()

	b, _ := svc.Create(ctx, "Dune", "H", "scifi", 5)
	if _, err := loans.Create(ctx, &model.Loan{BookID: b.ID, UserID: 1, Status: model.LoanActive}); err != nil {
		t.Fatalf("seed loan 1 failed: %v", err)
	}
	if _, err := loans.Create(ctx, &model.Loan{BookID: b.ID, UserID: 2, Status: model.LoanActive}); err != nil {
		t.Fatalf("seed loan 2 failed: %v", err)
	}

	err := svc.Update(ctx, b.ID, service.UpdateBookInput{Title: "Dune", Author: "H", Genre: "scifi", Copies: 1})
	if !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("want ErrInvalid, got %v", err)
	}
}

func TestBookService_Update_Success(t *testing.T) {
	books, loans := newFakeBookRepo(), newFakeLoanRepo()
	svc := service.NewBookService(books, loans)
	ctx := context.Background()

	b, _ := svc.Create(ctx, "Dune", "H", "scifi", 5)

	err := svc.Update(ctx, b.ID, service.UpdateBookInput{Title: "Dune 2", Author: "H", Genre: "scifi", Copies: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.Get(ctx, b.ID)
	if got.Title != "Dune 2" || got.Copies != 3 {
		t.Fatalf("update didn't apply: %+v", got)
	}
}

func TestBookService_Delete(t *testing.T) {
	svc := service.NewBookService(newFakeBookRepo(), newFakeLoanRepo())
	ctx := context.Background()

	b, _ := svc.Create(ctx, "Dune", "H", "scifi", 1)
	if err := svc.Delete(ctx, b.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.Get(ctx, b.ID)
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
}

func TestBookService_SetCoverPath(t *testing.T) {
	svc := service.NewBookService(newFakeBookRepo(), newFakeLoanRepo())
	ctx := context.Background()

	b, _ := svc.Create(ctx, "Dune", "H", "scifi", 1)
	if err := svc.SetCoverPath(ctx, b.ID, "uploads/covers/book_1.jpg"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.Get(ctx, b.ID)
	if got.CoverPath != "uploads/covers/book_1.jpg" {
		t.Fatalf("cover path not set: %+v", got)
	}
}

func TestBookService_List_FilterByGenre(t *testing.T) {
	svc := service.NewBookService(newFakeBookRepo(), newFakeLoanRepo())
	ctx := context.Background()

	if _, err := svc.Create(ctx, "Dune", "H", "scifi", 1); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := svc.Create(ctx, "Clean Code", "M", "tech", 1); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	items, total, err := svc.List(ctx, repository.BookFilter{Genre: "tech"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("want 1 tech book, got total=%d len=%d", total, len(items))
	}
}
