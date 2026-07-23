package model

import "errors"

var (
	ErrNotFound  = errors.New("not found")
	ErrInvalid   = errors.New("invalid input")
	ErrForbidden = errors.New("forbidden")
	ErrNoSeats   = errors.New("no available copies")
)
