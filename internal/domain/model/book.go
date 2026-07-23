package model

import "time"

type BookStatus string

const (
	BookAvailable BookStatus = "available"
	BookBorrowed  BookStatus = "borrowed"
	BookLost      BookStatus = "lost"
)

type Book struct {
	ID        int64
	Title     string
	Author    string
	Genre     string
	CoverPath string
	Status    BookStatus
	Copies    int
	CreatedAt time.Time
}
