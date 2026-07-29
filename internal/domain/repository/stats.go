package repository

import "context"

type StatsRepo interface {
	LibraryStats(ctx context.Context) (*LibraryStatsRow, error)
}

type LibraryStatsRow struct {
	BooksTotal       int
	LoansActive      int
	LoansOverdue     int
	ReservationsWait int
}
