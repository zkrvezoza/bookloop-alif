package service

import (
	"context"
	"errors"
	"testing"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
)

type mockBookRepo struct {
	createFn     func(context.Context, *model.Book) (*model.Book, error)
	getByIDFn    func(context.Context, int64) (*model.Book, error)
	updateFn     func(context.Context, *model.Book) error
	deleteFn     func(context.Context, int64) error
	listFn       func(context.Context, repository.BookFilter) ([]model.Book, int, error)
	decrCopiesFn func(context.Context, int64) error
	incrCopiesFn func(context.Context, int64) error
	markLostFn   func(context.Context, int64) error
}

func (m *mockBookRepo) GetByID(
	ctx context.Context,
	id int64,
) (*model.Book, error) {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}

	return m.getByIDFn(ctx, id)
}

func TestBookService_Get_OK(t *testing.T) {
	repo := &mockBookRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Book, error) {
			return &model.Book{
				ID:     id,
				Title:  "Dune",
				Author: "Frank Herbert",
				Copies: 3,
			}, nil
		},
	}

	svc := NewBookService(repo, nil)

	got, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Title != "Dune" {
		t.Fatalf("expected Dune, got %q", got.Title)
	}
}

func TestBookService_Get_RepoError(t *testing.T) {
	expectedErr := errors.New("database failed")

	repo := &mockBookRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Book, error) {
			return nil, expectedErr
		},
	}

	svc := NewBookService(repo, nil)

	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func (m *mockBookRepo) Create(
	ctx context.Context,
	book *model.Book,
) (*model.Book, error) {
	if m.createFn == nil {
		panic("unexpected call to Create")
	}

	return m.createFn(ctx, book)
}

func (m *mockBookRepo) Update(
	ctx context.Context,
	book *model.Book,
) error {
	if m.updateFn == nil {
		panic("unexpected call to Update")
	}

	return m.updateFn(ctx, book)
}

func (m *mockBookRepo) Delete(
	ctx context.Context,
	id int64,
) error {
	if m.deleteFn == nil {
		panic("unexpected call to Delete")
	}

	return m.deleteFn(ctx, id)
}

func (m *mockBookRepo) List(
	ctx context.Context,
	filter repository.BookFilter,
) ([]model.Book, int, error) {
	if m.listFn == nil {
		panic("unexpected call to List")
	}

	return m.listFn(ctx, filter)
}

func (m *mockBookRepo) DecrCopies(
	ctx context.Context,
	id int64,
) error {
	if m.decrCopiesFn == nil {
		panic("unexpected call to DecrCopies")
	}

	return m.decrCopiesFn(ctx, id)
}

func (m *mockBookRepo) IncrCopies(
	ctx context.Context,
	id int64,
) error {
	if m.incrCopiesFn == nil {
		panic("unexpected call to IncrCopies")
	}

	return m.incrCopiesFn(ctx, id)
}

func (m *mockBookRepo) MarkLost(
	ctx context.Context,
	id int64,
) error {
	if m.markLostFn == nil {
		panic("unexpected call to MarkLost")
	}

	return m.markLostFn(ctx, id)
}

func TestBookService_Create_OK(t *testing.T) {
	repo := &mockBookRepo{
		createFn: func(ctx context.Context, book *model.Book) (*model.Book, error) {
			book.ID = 1
			return book, nil
		},
	}

	svc := NewBookService(repo, nil)

	got, err := svc.Create(
		context.Background(),
		"Dune",
		"Frank Herbert",
		"Science Fiction",
		3,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Title != "Dune" {
		t.Fatalf("expected title Dune, got %q", got.Title)
	}

	if got.Status != model.BookAvailable {
		t.Fatalf("expected status %q, got %q", model.BookAvailable, got.Status)
	}

	if got.Copies != 3 {
		t.Fatalf("expected 3 copies, got %d", got.Copies)
	}
}

func TestBookService_Create_InvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		title  string
		author string
		genre  string
		copies int
	}{
		{
			name:   "empty title",
			title:  "",
			author: "Frank Herbert",
			genre:  "Science Fiction",
			copies: 3,
		},
		{
			name:   "empty author",
			title:  "Dune",
			author: "",
			genre:  "Science Fiction",
			copies: 3,
		},
		{
			name:   "empty genre",
			title:  "Dune",
			author: "Frank Herbert",
			genre:  "",
			copies: 3,
		},
		{
			name:   "zero copies",
			title:  "Dune",
			author: "Frank Herbert",
			genre:  "Science Fiction",
			copies: 0,
		},
		{
			name:   "negative copies",
			title:  "Dune",
			author: "Frank Herbert",
			genre:  "Science Fiction",
			copies: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewBookService(&mockBookRepo{}, nil)

			got, err := svc.Create(
				context.Background(),
				tt.title,
				tt.author,
				tt.genre,
				tt.copies,
			)

			if !errors.Is(err, model.ErrInvalid) {
				t.Fatalf("expected ErrInvalid, got %v", err)
			}

			if got != nil {
				t.Fatalf("expected nil book, got %#v", got)
			}
		})
	}
}

func TestBookService_Create_RepoError(t *testing.T) {
	expectedErr := errors.New("database failed")

	repo := &mockBookRepo{
		createFn: func(ctx context.Context, book *model.Book) (*model.Book, error) {
			return nil, expectedErr
		},
	}

	svc := NewBookService(repo, nil)

	_, err := svc.Create(
		context.Background(),
		"Dune",
		"Frank Herbert",
		"Science Fiction",
		3,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestBookService_List_OK(t *testing.T) {
	expectedBooks := []model.Book{
		{ID: 1, Title: "Dune"},
		{ID: 2, Title: "1984"},
	}

	repo := &mockBookRepo{
		listFn: func(
			ctx context.Context,
			filter repository.BookFilter,
		) ([]model.Book, int, error) {
			return expectedBooks, 2, nil
		},
	}

	svc := NewBookService(repo, nil)

	got, total, err := svc.List(
		context.Background(),
		repository.BookFilter{Page: 1, Limit: 10},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 books, got %d", len(got))
	}
}

func TestBookService_Delete_OK(t *testing.T) {
	var deletedID int64

	repo := &mockBookRepo{
		deleteFn: func(ctx context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	svc := NewBookService(repo, nil)

	err := svc.Delete(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deletedID != 7 {
		t.Fatalf("expected deleted ID 7, got %d", deletedID)
	}
}

func TestBookService_Delete_RepoError(t *testing.T) {
	expectedErr := errors.New("delete failed")

	repo := &mockBookRepo{
		deleteFn: func(ctx context.Context, id int64) error {
			return expectedErr
		},
	}

	svc := NewBookService(repo, nil)

	err := svc.Delete(context.Background(), 7)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
