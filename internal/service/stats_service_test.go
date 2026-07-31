package service

import (
	"context"
	"errors"
	"testing"

	"github.com/bookloop-alif/internal/domain/repository"
)

type mockStatsRepo struct {
	libraryStatsFn func(context.Context) (*repository.LibraryStatsRow, error)
}

func (m *mockStatsRepo) LibraryStats(
	ctx context.Context,
) (*repository.LibraryStatsRow, error) {
	if m.libraryStatsFn == nil {
		panic("unexpected call")
	}

	return m.libraryStatsFn(ctx)
}

func TestStatsService_Get_OK(t *testing.T) {
	repo := &mockStatsRepo{
		libraryStatsFn: func(
			ctx context.Context,
		) (*repository.LibraryStatsRow, error) {

			return &repository.LibraryStatsRow{
				BooksTotal:       120,
				LoansActive:      14,
				LoansOverdue:     3,
				ReservationsWait: 8,
			}, nil
		},
	}

	svc := NewStatsService(repo)

	got, err := svc.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if got.BooksTotal != 120 {
		t.Fatal()
	}

	if got.LoansActive != 14 {
		t.Fatal()
	}

	if got.LoansOverdue != 3 {
		t.Fatal()
	}

	if got.ReservationsWait != 8 {
		t.Fatal()
	}
}

func TestStatsService_Get_Error(t *testing.T) {
	expected := errors.New("db failed")

	repo := &mockStatsRepo{
		libraryStatsFn: func(
			ctx context.Context,
		) (*repository.LibraryStatsRow, error) {
			return nil, expected
		},
	}

	svc := NewStatsService(repo)

	_, err := svc.Get(context.Background())

	if !errors.Is(err, expected) {
		t.Fatal(err)
	}
}
