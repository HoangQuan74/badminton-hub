package model

import "time"

type Role struct {
	ID, Name, Description string
	CreatedAt             time.Time
}

type Permission struct {
	ID, Module, Path, Action, Description string
}

type StaffCredentials struct {
	Staff        Staff
	PasswordHash string
}
