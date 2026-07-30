package service_test

import (
	"context"
	"testing"

	"github.com/bookloop-alif/internal/domain/repository"
	"github.com/bookloop-alif/internal/service"
)

type fakeStatsRepo struct {
	row *repository.LibraryStatsRow
}

func (f *fakeStatsRepo) LibraryStats(_ context.Context) (*repository.LibraryStatsRow, error) {
	return f.row, nil
}

func TestStatsService_Get(t *testing.T) {
	fake := &fakeStatsRepo{row: &repository.LibraryStatsRow{
		BooksTotal: 10, LoansActive: 3, LoansOverdue: 1, ReservationsWait: 2,
	}}
	svc := service.NewStatsService(fake)

	stats, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.BooksTotal != 10 || stats.LoansActive != 3 || stats.LoansOverdue != 1 || stats.ReservationsWait != 2 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}
