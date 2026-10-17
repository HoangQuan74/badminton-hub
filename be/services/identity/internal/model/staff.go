package model

import "time"

type Staff struct {
	ID        string
	Email     string
	FullName  string
	IsAdmin   bool
	RoleID    string
	RoleName  string
	Status    string
	CreatedAt time.Time
}
