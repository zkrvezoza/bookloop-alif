package postgres

import (
	"context"

	"github.com/bookloop-alif/internal/domain/repository"
)

type UnitOfWork struct {
	tx *TxManager
}

func NewUnitOfWork(tx *TxManager) *UnitOfWork {
	return &UnitOfWork{tx: tx}
}

func (u *UnitOfWork) WithinTx(ctx context.Context, fn func(books repository.BookRepo, loans repository.LoanRepo) error) error {
	return u.tx.WithTx(ctx, func(q Querier) error {
		books := NewBookRepo(q)
		loans := NewLoanRepo(q)
		return fn(books, loans)
	})
}
