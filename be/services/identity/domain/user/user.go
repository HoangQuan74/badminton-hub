package user

import "time"

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
