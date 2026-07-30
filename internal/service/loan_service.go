package service

import (
	"context"
	"errors"
	"time"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
)

type LoanService struct {
	loans repository.LoanRepo
	books repository.BookRepo
	res   repository.ReservationRepo
}

func NewLoanService(loans repository.LoanRepo, books repository.BookRepo, res repository.ReservationRepo) *LoanService {
	return &LoanService{loans: loans, books: books, res: res}
}

func (s *LoanService) List(ctx context.Context, f repository.LoanFilter) ([]model.Loan, int, error) {
	return s.loans.List(ctx, f)
}

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
		if errors.Is(err, model.ErrNotFound) {
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

func (s *LoanService) MarkLost(ctx context.Context, bookID int64) error {
	activeLoans, _, err := s.loans.List(ctx, repository.LoanFilter{Status: string(model.LoanActive)})
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
