package service_test

import (
	"context"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/domain/repository"
)

type fakeBookRepo struct {
	books  map[int64]*model.Book
	nextID int64
}

func newFakeBookRepo() *fakeBookRepo {
	return &fakeBookRepo{books: make(map[int64]*model.Book)}
}

func (f *fakeBookRepo) Create(_ context.Context, b *model.Book) (*model.Book, error) {
	f.nextID++
	b.ID = f.nextID
	f.books[b.ID] = b
	return b, nil
}

func (f *fakeBookRepo) GetByID(_ context.Context, id int64) (*model.Book, error) {
	b, ok := f.books[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return b, nil
}

func (f *fakeBookRepo) Update(_ context.Context, b *model.Book) error {
	if _, ok := f.books[b.ID]; !ok {
		return model.ErrNotFound
	}
	f.books[b.ID] = b
	return nil
}

func (f *fakeBookRepo) Delete(_ context.Context, id int64) error {
	if _, ok := f.books[id]; !ok {
		return model.ErrNotFound
	}
	delete(f.books, id)
	return nil
}

func (f *fakeBookRepo) List(_ context.Context, filter repository.BookFilter) ([]model.Book, int, error) {
	var out []model.Book
	for _, b := range f.books {
		if filter.Genre != "" && b.Genre != filter.Genre {
			continue
		}
		out = append(out, *b)
	}
	return out, len(out), nil
}

func (f *fakeBookRepo) DecrCopies(_ context.Context, id int64) error {
	b, ok := f.books[id]
	if !ok {
		return model.ErrNotFound
	}
	if b.Copies <= 0 {
		return model.ErrNoSeats
	}
	b.Copies--
	return nil
}

func (f *fakeBookRepo) IncrCopies(_ context.Context, id int64) error {
	b, ok := f.books[id]
	if !ok {
		return model.ErrNotFound
	}
	b.Copies++
	return nil
}

func (f *fakeBookRepo) MarkLost(_ context.Context, id int64) error {
	b, ok := f.books[id]
	if !ok {
		return model.ErrNotFound
	}
	b.Status = model.BookLost
	return nil
}

type fakeLoanRepo struct {
	loans  map[int64]*model.Loan
	nextID int64
}

func newFakeLoanRepo() *fakeLoanRepo {
	return &fakeLoanRepo{loans: make(map[int64]*model.Loan)}
}

func (f *fakeLoanRepo) Create(_ context.Context, l *model.Loan) (*model.Loan, error) {
	f.nextID++
	l.ID = f.nextID
	f.loans[l.ID] = l
	return l, nil
}

func (f *fakeLoanRepo) GetByID(_ context.Context, id int64) (*model.Loan, error) {
	l, ok := f.loans[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return l, nil
}

func (f *fakeLoanRepo) List(_ context.Context, filter repository.LoanFilter) ([]model.Loan, int, error) {
	var out []model.Loan
	for _, l := range f.loans {
		if filter.UserID != 0 && l.UserID != filter.UserID {
			continue
		}
		if filter.Status != "" && string(l.Status) != filter.Status {
			continue
		}
		out = append(out, *l)
	}
	return out, len(out), nil
}

func (f *fakeLoanRepo) MarkReturned(_ context.Context, id int64) error {
	l, ok := f.loans[id]
	if !ok || l.Status != model.LoanActive {
		return model.ErrNotFound
	}
	l.Status = model.LoanReturned
	return nil
}

func (f *fakeLoanRepo) MarkOverdue(_ context.Context, id int64) error {
	l, ok := f.loans[id]
	if !ok || l.Status != model.LoanActive {
		return model.ErrNotFound
	}
	l.Status = model.LoanOverdue
	return nil
}

func (f *fakeLoanRepo) ExtendDueDate(_ context.Context, id int64, _ string) error {
	l, ok := f.loans[id]
	if !ok || l.Status != model.LoanActive {
		return model.ErrNotFound
	}
	return nil
}

func (f *fakeLoanRepo) CountActiveByBook(_ context.Context, bookID int64) (int, error) {
	count := 0
	for _, l := range f.loans {
		if l.BookID == bookID && l.Status == model.LoanActive {
			count++
		}
	}
	return count, nil
}

type fakeReservationRepo struct {
	res    map[int64]*model.Reservation
	nextID int64
}

func newFakeReservationRepo() *fakeReservationRepo {
	return &fakeReservationRepo{res: make(map[int64]*model.Reservation)}
}

func (f *fakeReservationRepo) Create(_ context.Context, r *model.Reservation) (*model.Reservation, error) {
	f.nextID++
	r.ID = f.nextID
	f.res[r.ID] = r
	return r, nil
}

func (f *fakeReservationRepo) NextWaiting(_ context.Context, bookID int64) (*model.Reservation, error) {
	var found *model.Reservation
	for _, r := range f.res {
		if r.BookID == bookID && r.Status == model.ReservationWaiting {
			if found == nil || r.ID < found.ID {
				found = r
			}
		}
	}
	if found == nil {
		return nil, model.ErrNotFound
	}
	return found, nil
}

func (f *fakeReservationRepo) MarkFulfilled(_ context.Context, id int64) error {
	r, ok := f.res[id]
	if !ok || r.Status != model.ReservationWaiting {
		return model.ErrNotFound
	}
	r.Status = model.ReservationFulfilled
	return nil
}

func (f *fakeReservationRepo) Cancel(_ context.Context, id int64) error {
	r, ok := f.res[id]
	if !ok || r.Status != model.ReservationWaiting {
		return model.ErrNotFound
	}
	r.Status = model.ReservationCancelled
	return nil
}

func (f *fakeReservationRepo) ListByUser(_ context.Context, userID int64) ([]model.Reservation, error) {
	var out []model.Reservation
	for _, r := range f.res {
		if r.UserID == userID {
			out = append(out, *r)
		}
	}
	return out, nil
}
