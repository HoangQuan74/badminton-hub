package auth

import (
	"context"
	"time"

	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type CreateUserParams struct {
	ID           string
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
}

type Session struct {
	ID        string
	UserID    string
	TokenHash []byte
	ExpiresAt time.Time
}

type Repository interface {
	CreateUserWithSession(context.Context, CreateUserParams, Session) (model.User, error)
	FindUserByEmail(context.Context, string) (model.User, error)
	FindUserByID(context.Context, string) (model.User, error)
	CreateSession(context.Context, Session) error
	RotateSession(context.Context, []byte, Session, time.Time) (string, error)
	RevokeSession(context.Context, []byte, time.Time) error
}
