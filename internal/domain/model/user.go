package model

import "time"

type Role string

const (
	RoleUser      Role = "user"
	RoleLibrarian Role = "librarian"
)

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
}
