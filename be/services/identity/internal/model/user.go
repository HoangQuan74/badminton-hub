package model

import "time"

type AccountStatus string

type User struct {
	ID           string
	Email        string
	PasswordHash string `json:"-"`
	FullName     string
	Phone        string
	Status       AccountStatus
	RoleName     string
	IsAdmin      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
