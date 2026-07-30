package service

import (
	"context"

	"github.com/bookloop-alif/internal/domain/repository"
)

type StatsService struct {
	stats repository.StatsRepo
}

func NewStatsService(stats repository.StatsRepo) *StatsService {
	return &StatsService{stats: stats}
}

type LibraryStats struct {
	BooksTotal       int `json:"books_total"`
	LoansActive      int `json:"loans_active"`
	LoansOverdue     int `json:"loans_overdue"`
	ReservationsWait int `json:"reservations_waiting"`
}

func (s *StatsService) Get(ctx context.Context) (*LibraryStats, error) {
	row, err := s.stats.LibraryStats(ctx)
	if err != nil {
		return nil, err
	}
	return &LibraryStats{
		BooksTotal:       row.BooksTotal,
		LoansActive:      row.LoansActive,
		LoansOverdue:     row.LoansOverdue,
		ReservationsWait: row.ReservationsWait,
	}, nil
}
