package repository

import (
	"context"

	"github.com/bookloop-alif/internal/domain/model"
)

type BookFilter struct {
	Query  string
	Genre  string
	Status string
	Page   int
	Limit  int
}

type BookRepo interface {
	Create(ctx context.Context, b *model.Book) (*model.Book, error)
	GetByID(ctx context.Context, id int64) (*model.Book, error)
	Update(ctx context.Context, b *model.Book) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, f BookFilter) ([]model.Book, int, error)
	DecrCopies(ctx context.Context, id int64) error
	IncrCopies(ctx context.Context, id int64) error
	MarkLost(ctx context.Context, id int64) error
}
