package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/service"
)

func TestLoanService_Return_NoQueue(t *testing.T) {
	books, loans, res := newFakeBookRepo(), newFakeLoanRepo(), newFakeReservationRepo()
	svc := service.NewLoanService(loans, books, res)
	ctx := context.Background()

	b, _ := books.Create(ctx, &model.Book{Title: "T", Copies: 0, Status: model.BookAvailable})
	l, _ := loans.Create(ctx, &model.Loan{BookID: b.ID, UserID: 1, DueAt: time.Now(), Status: model.LoanActive})

	if err := svc.Return(ctx, l.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := loans.GetByID(ctx, l.ID)
	if got.Status != model.LoanReturned {
		t.Fatalf("want returned, got %s", got.Status)
	}
	gotBook, _ := books.GetByID(ctx, b.ID)
	if gotBook.Copies != 1 {
		t.Fatalf("want copies incremented to 1, got %d", gotBook.Copies)
	}
}

func TestLoanService_Return_PromotesQueue(t *testing.T) {
	books, loans, res := newFakeBookRepo(), newFakeLoanRepo(), newFakeReservationRepo()
	svc := service.NewLoanService(loans, books, res)
	ctx := context.Background()

	b, _ := books.Create(ctx, &model.Book{Title: "T", Copies: 0, Status: model.BookAvailable})
	l, _ := loans.Create(ctx, &model.Loan{BookID: b.ID, UserID: 1, DueAt: time.Now(), Status: model.LoanActive})
	r, _ := res.Create(ctx, &model.Reservation{BookID: b.ID, UserID: 2, Status: model.ReservationWaiting})

	if err := svc.Return(ctx, l.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotRes, _ := res.res[r.ID], error(nil)
	if gotRes.Status != model.ReservationFulfilled {
		t.Fatalf("want reservation fulfilled, got %s", gotRes.Status)
	}

	// новый loan должен быть создан для юзера из очереди
	allLoans, _, _ := loans.List(ctx, repository.LoanFilter{UserID: 2})
	if len(allLoans) != 1 {
		t.Fatalf("want 1 new loan for queued user, got %d", len(allLoans))
	}
}

func TestLoanService_MarkLost(t *testing.T) {
	books, loans, res := newFakeBookRepo(), newFakeLoanRepo(), newFakeReservationRepo()
	svc := service.NewLoanService(loans, books, res)
	ctx := context.Background()

	b, _ := books.Create(ctx, &model.Book{Title: "T", Copies: 1, Status: model.BookAvailable})
	l, _ := loans.Create(ctx, &model.Loan{BookID: b.ID, UserID: 1, DueAt: time.Now(), Status: model.LoanActive})

	if err := svc.MarkLost(ctx, b.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotLoan, _ := loans.GetByID(ctx, l.ID)
	if gotLoan.Status != model.LoanReturned {
		t.Fatalf("want loan closed, got %s", gotLoan.Status)
	}
	gotBook, _ := books.GetByID(ctx, b.ID)
	if gotBook.Status != model.BookLost {
		t.Fatalf("want book status lost, got %s", gotBook.Status)
	}
}

func TestLoanService_ExtendDueDate(t *testing.T) {
	books, loans, res := newFakeBookRepo(), newFakeLoanRepo(), newFakeReservationRepo()
	svc := service.NewLoanService(loans, books, res)
	ctx := context.Background()

	l, _ := loans.Create(ctx, &model.Loan{BookID: 1, UserID: 1, DueAt: time.Now(), Status: model.LoanActive})

	if err := svc.ExtendDueDate(ctx, l.ID, 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoanService_ExtendDueDate_RejectsNonActive(t *testing.T) {
	books, loans, res := newFakeBookRepo(), newFakeLoanRepo(), newFakeReservationRepo()
	svc := service.NewLoanService(loans, books, res)
	ctx := context.Background()

	l, _ := loans.Create(ctx, &model.Loan{BookID: 1, UserID: 1, DueAt: time.Now(), Status: model.LoanActive})
	if err := loans.MarkReturned(ctx, l.ID); err != nil {
		t.Fatalf("mark returned failed: %v", err)
	}

	err := svc.ExtendDueDate(ctx, l.ID, 7)
	if !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("want ErrInvalid, got %v", err)
	}
}

func TestLoanService_List(t *testing.T) {
	books, loans, res := newFakeBookRepo(), newFakeLoanRepo(), newFakeReservationRepo()
	svc := service.NewLoanService(loans, books, res)
	ctx := context.Background()

	if _, err := loans.Create(ctx, &model.Loan{BookID: 1, UserID: 1, DueAt: time.Now(), Status: model.LoanActive}); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	items, total, err := svc.List(ctx, repository.LoanFilter{UserID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("want 1 loan for user 1, got total=%d len=%d", total, len(items))
	}
}
