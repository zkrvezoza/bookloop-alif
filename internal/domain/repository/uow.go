package repository

import "context"

type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(books BookRepo, loans LoanRepo) error) error
}
