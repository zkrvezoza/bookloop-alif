package repository

import (
	"context"

	"github.com/bookloop-alif/internal/domain/model"
)

type LoanFilter struct {
	UserID int64
	Status string
	Page   int
	Limit  int
}

type LoanRepo interface {
	Create(ctx context.Context, l *model.Loan) (*model.Loan, error)
	GetByID(ctx context.Context, id int64) (*model.Loan, error)
	List(ctx context.Context, f LoanFilter) ([]model.Loan, int, error)
	MarkReturned(ctx context.Context, id int64) error
	MarkOverdue(ctx context.Context, id int64) error
	ExtendDueDate(ctx context.Context, id int64, newDueAt string) error
	CountActiveByBook(ctx context.Context, bookID int64) (int, error)
}
