package postgres

import (
	"context"

	"github.com/bookloop-alif/internal/domain/repository"
)

type StatsRepo struct {
	db Querier
}

func NewStatsRepo(db Querier) *StatsRepo {
	return &StatsRepo{db: db}
}

// LibraryStats — один SQL-запрос с подзапросами-агрегатами (books/loans/reservations).
func (r *StatsRepo) LibraryStats(ctx context.Context) (*repository.LibraryStatsRow, error) {
	const q = `
		SELECT
			(SELECT COUNT(*) FROM books) AS books_total,
			(SELECT COUNT(*) FROM loans WHERE status = 'active') AS loans_active,
			(SELECT COUNT(*) FROM loans WHERE status = 'overdue') AS loans_overdue,
			(SELECT COUNT(*) FROM reservations WHERE status = 'waiting') AS reservations_waiting`

	row := &repository.LibraryStatsRow{}
	err := r.db.QueryRow(ctx, q).
		Scan(&row.BooksTotal, &row.LoansActive, &row.LoansOverdue, &row.ReservationsWait)
	if err != nil {
		return nil, err
	}
	return row, nil
}
