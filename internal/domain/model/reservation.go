package model

import "time"

type ReservationStatus string

const (
	ReservationWaiting   ReservationStatus = "waiting"
	ReservationFulfilled ReservationStatus = "fulfilled"
	ReservationCancelled ReservationStatus = "cancelled"
)

type Reservation struct {
	ID        int64
	BookID    int64
	UserID    int64
	CreatedAt time.Time
	Status    ReservationStatus
}
