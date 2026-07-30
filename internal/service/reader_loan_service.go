package service

import (
	"context"
	"time"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/repository/postgres"
)

const loanPeriod = 14 * 24 * time.Hour

type ReaderLoanService struct {
	txManager *postgres.TxManager
	loans     repository.LoanRepo
	books     repository.BookRepo
	res       repository.ReservationRepo
	shared    *LoanService
}

func NewReaderLoanService(
	tx *postgres.TxManager,
	loans repository.LoanRepo,
	books repository.BookRepo,
	res repository.ReservationRepo,
	shared *LoanService,
) *ReaderLoanService {
	return &ReaderLoanService{txManager: tx, loans: loans, books: books, res: res, shared: shared}
}

// Borrow — A-02.
func (s *ReaderLoanService) Borrow(ctx context.Context, userID, bookID int64) (*model.Loan, error) {
	var loan *model.Loan

	err := s.txManager.WithTx(ctx, func(q postgres.Querier) error {
		bookRepo := postgres.NewBookRepo(q)
		loanRepo := postgres.NewLoanRepo(q)

		if err := bookRepo.DecrCopies(ctx, bookID); err != nil {
			return err
		}

		l, err := loanRepo.Create(ctx, &model.Loan{
			BookID: bookID,
			UserID: userID,
			DueAt:  time.Now().Add(loanPeriod),
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

// MyLoans — A-03.
func (s *ReaderLoanService) MyLoans(ctx context.Context, userID int64, f repository.LoanFilter) ([]model.Loan, int, error) {
	f.UserID = userID
	return s.loans.List(ctx, f)
}

// Return — A-03.
func (s *ReaderLoanService) Return(ctx context.Context, userID, loanID int64) error {
	loan, err := s.loans.GetByID(ctx, loanID)
	if err != nil {
		return err
	}
	if loan.UserID != userID {
		return model.ErrForbidden
	}
	return s.shared.Return(ctx, loanID)
}
