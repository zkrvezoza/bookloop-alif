package service

import (
	"context"
	"time"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/repository/postgres"
)

type LoanService struct {
	loans repository.LoanRepo
	books repository.BookRepo
	res   repository.ReservationRepo
	tx    *postgres.TxManager
}

func NewLoanService(loans repository.LoanRepo, books repository.BookRepo, res repository.ReservationRepo, tx *postgres.TxManager) *LoanService {
	return &LoanService{loans: loans, books: books, res: res, tx: tx}
}

// List — B-02: все выдачи с фильтрами.
func (s *LoanService) List(ctx context.Context, f repository.LoanFilter) ([]model.Loan, int, error) {
	return s.loans.List(ctx, f)
}

// Borrow — A-02: взять книгу. Атомарно через TxManager.
func (s *LoanService) Borrow(ctx context.Context, bookID, userID int64) (*model.Loan, error) {
	var loan *model.Loan

	err := s.tx.WithTx(ctx, func(q postgres.Querier) error {
		bookRepo := postgres.NewBookRepo(q)
		loanRepo := postgres.NewLoanRepo(q)

		if err := bookRepo.DecrCopies(ctx, bookID); err != nil {
			return err
		}

		l, err := loanRepo.Create(ctx, &model.Loan{
			BookID: bookID,
			UserID: userID,
			DueAt:  time.Now().Add(14 * 24 * time.Hour),
			Status: model.LoanActive,
		})
		if err != nil {
			return err
		}
		loan = l
		return nil
	})

	return loan, err
}

// MyLoans — A-03: свои выдачи читателя.
func (s *LoanService) MyLoans(ctx context.Context, userID int64) ([]model.Loan, error) {
	loans, _, err := s.loans.List(ctx, repository.LoanFilter{UserID: userID})
	return loans, err
}

// Return — B-03: возврат (за читателя или библиотекаря — единая логика). Авто-выдача из очереди.
func (s *LoanService) Return(ctx context.Context, loanID int64) error {
	loan, err := s.loans.GetByID(ctx, loanID)
	if err != nil {
		return err
	}

	if err := s.loans.MarkReturned(ctx, loanID); err != nil {
		return err
	}
	if err := s.books.IncrCopies(ctx, loan.BookID); err != nil {
		return err
	}

	next, err := s.res.NextWaiting(ctx, loan.BookID)
	if err != nil {
		if err == model.ErrNotFound {
			return nil
		}
		return err
	}

	if err := s.res.MarkFulfilled(ctx, next.ID); err != nil {
		return err
	}
	if err := s.books.DecrCopies(ctx, loan.BookID); err != nil {
		return err
	}

	_, err = s.loans.Create(ctx, &model.Loan{
		BookID: loan.BookID,
		UserID: next.UserID,
		DueAt:  time.Now().Add(14 * 24 * time.Hour),
		Status: model.LoanActive,
	})
	return err
}

func (s *LoanService) ReturnOwn(ctx context.Context, loanID, userID int64) error {
	loan, err := s.loans.GetByID(ctx, loanID)
	if err != nil {
		return err
	}
	if loan.UserID != userID {
		return model.ErrForbidden
	}
	return s.Return(ctx, loanID)
}

func (s *LoanService) MarkLost(ctx context.Context, bookID int64) error {
	activeLoans, err := s.loans.List(ctx, repository.LoanFilter{Status: string(model.LoanActive)})
	if err != nil {
		return err
	}
	for _, l := range activeLoans {
		if l.BookID == bookID {
			if err := s.loans.MarkReturned(ctx, l.ID); err != nil {
				return err
			}
		}
	}
	return s.books.MarkLost(ctx, bookID)
}

func (s *LoanService) ExtendDueDate(ctx context.Context, loanID int64, days int) error {
	loan, err := s.loans.GetByID(ctx, loanID)
	if err != nil {
		return err
	}
	if loan.Status != model.LoanActive {
		return model.ErrInvalid
	}
	newDue := loan.DueAt.Add(time.Duration(days) * 24 * time.Hour)
	return s.loans.ExtendDueDate(ctx, loanID, newDue.Format(time.RFC3339))
}
