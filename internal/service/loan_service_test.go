package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
)

type mockLoanRepo struct {
	createFn            func(context.Context, *model.Loan) (*model.Loan, error)
	getByIDFn           func(context.Context, int64) (*model.Loan, error)
	listFn              func(context.Context, repository.LoanFilter) ([]model.Loan, int, error)
	markReturnedFn      func(context.Context, int64) error
	markOverdueFn       func(context.Context, int64) error
	extendDueDateFn     func(context.Context, int64, string) error
	countActiveByBookFn func(context.Context, int64) (int, error)
}

func (m *mockLoanRepo) Create(
	ctx context.Context,
	loan *model.Loan,
) (*model.Loan, error) {
	if m.createFn == nil {
		panic("unexpected call to Create")
	}

	return m.createFn(ctx, loan)
}

func (m *mockLoanRepo) GetByID(
	ctx context.Context,
	loanID int64,
) (*model.Loan, error) {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}

	return m.getByIDFn(ctx, loanID)
}

func (m *mockLoanRepo) List(
	ctx context.Context,
	filter repository.LoanFilter,
) ([]model.Loan, int, error) {
	if m.listFn == nil {
		panic("unexpected call to List")
	}

	return m.listFn(ctx, filter)
}

func (m *mockLoanRepo) MarkReturned(
	ctx context.Context,
	loanID int64,
) error {
	if m.markReturnedFn == nil {
		panic("unexpected call to MarkReturned")
	}

	return m.markReturnedFn(ctx, loanID)
}

func (m *mockLoanRepo) ExtendDueDate(
	ctx context.Context,
	loanID int64,
	dueAt string,
) error {
	if m.extendDueDateFn == nil {
		panic("unexpected call to ExtendDueDate")
	}

	return m.extendDueDateFn(ctx, loanID, dueAt)
}
func (m *mockLoanRepo) MarkOverdue(
	ctx context.Context,
	id int64,
) error {
	if m.markOverdueFn == nil {
		panic("unexpected call to MarkOverdue")
	}

	return m.markOverdueFn(ctx, id)
}

func (m *mockLoanRepo) CountActiveByBook(
	ctx context.Context,
	bookID int64,
) (int, error) {
	if m.countActiveByBookFn == nil {
		panic("unexpected call to CountActiveByBook")
	}

	return m.countActiveByBookFn(ctx, bookID)
}

func TestLoanService_List_OK(t *testing.T) {
	loanRepo := &mockLoanRepo{
		listFn: func(
			ctx context.Context,
			filter repository.LoanFilter,
		) ([]model.Loan, int, error) {
			return []model.Loan{
				{ID: 1, UserID: 2, BookID: 3},
			}, 1, nil
		},
	}

	svc := NewLoanService(loanRepo, nil, nil, nil)

	got, total, err := svc.List(
		context.Background(),
		repository.LoanFilter{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}

	if len(got) != 1 {
		t.Fatalf("expected one loan, got %d", len(got))
	}
}

func TestLoanService_ExtendDueDate_OK(t *testing.T) {
	initialDue := time.Now().UTC().Truncate(time.Second)
	var receivedDue string

	loanRepo := &mockLoanRepo{
		getByIDFn: func(
			ctx context.Context,
			loanID int64,
		) (*model.Loan, error) {
			return &model.Loan{
				ID:     loanID,
				DueAt:  initialDue,
				Status: model.LoanActive,
			}, nil
		},
		extendDueDateFn: func(
			ctx context.Context,
			loanID int64,
			dueAt string,
		) error {
			receivedDue = dueAt
			return nil
		},
	}

	svc := NewLoanService(loanRepo, nil, nil, nil)

	err := svc.ExtendDueDate(context.Background(), 5, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDue := initialDue.Add(7 * 24 * time.Hour).Format(time.RFC3339)

	if receivedDue != expectedDue {
		t.Fatalf(
			"expected due date %q, got %q",
			expectedDue,
			receivedDue,
		)
	}
}

func TestLoanService_ExtendDueDate_GetError(t *testing.T) {
	expectedErr := errors.New("loan not found")

	loanRepo := &mockLoanRepo{
		getByIDFn: func(
			ctx context.Context,
			loanID int64,
		) (*model.Loan, error) {
			return nil, expectedErr
		},
	}

	svc := NewLoanService(loanRepo, nil, nil, nil)

	err := svc.ExtendDueDate(context.Background(), 5, 7)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestLoanService_ExtendDueDate_InactiveLoan(t *testing.T) {
	loanRepo := &mockLoanRepo{
		getByIDFn: func(
			ctx context.Context,
			loanID int64,
		) (*model.Loan, error) {
			return &model.Loan{
				ID:     loanID,
				Status: model.LoanReturned,
			}, nil
		},
	}

	svc := NewLoanService(loanRepo, nil, nil, nil)

	err := svc.ExtendDueDate(context.Background(), 5, 7)
	if !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestLoanService_MyLoans_OK(t *testing.T) {
	loanRepo := &mockLoanRepo{
		listFn: func(
			ctx context.Context,
			filter repository.LoanFilter,
		) ([]model.Loan, int, error) {
			if filter.UserID != 77 {
				t.Fatalf(
					"expected user ID 77, got %d",
					filter.UserID,
				)
			}

			return []model.Loan{
				{ID: 1, UserID: 77},
			}, 1, nil
		},
	}

	svc := NewLoanService(loanRepo, nil, nil, nil)

	got, total, err := svc.MyLoans(
		context.Background(),
		77,
		repository.LoanFilter{
			Status: string(model.LoanActive),
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if total != 1 || len(got) != 1 {
		t.Fatalf(
			"expected one loan and total 1, got len=%d total=%d",
			len(got),
			total,
		)
	}
}

func TestLoanService_ReturnOwn_Forbidden(t *testing.T) {
	loanRepo := &mockLoanRepo{
		getByIDFn: func(
			ctx context.Context,
			loanID int64,
		) (*model.Loan, error) {
			return &model.Loan{
				ID:     loanID,
				UserID: 100,
			}, nil
		},
	}

	svc := NewLoanService(loanRepo, nil, nil, nil)

	err := svc.ReturnOwn(context.Background(), 200, 5)

	if !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestLoanService_ReturnOwn_GetError(t *testing.T) {
	expectedErr := errors.New("database failed")

	loanRepo := &mockLoanRepo{
		getByIDFn: func(
			ctx context.Context,
			loanID int64,
		) (*model.Loan, error) {
			return nil, expectedErr
		},
	}

	svc := NewLoanService(loanRepo, nil, nil, nil)

	err := svc.ReturnOwn(context.Background(), 200, 5)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
