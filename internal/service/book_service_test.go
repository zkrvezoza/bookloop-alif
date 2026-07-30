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
