package model

import "time"

type LoanStatus string

const (
	LoanActive   LoanStatus = "active"
	LoanReturned LoanStatus = "returned"
	LoanOverdue  LoanStatus = "overdue"
)

type Loan struct {
	ID         int64
	BookID     int64
	UserID     int64
	BorrowedAt time.Time
	DueAt      time.Time
	ReturnedAt *time.Time // nil, пока не вернули
	Status     LoanStatus
}
