package auth

import (
	"context"
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/domain/user"
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
	CreateUserWithSession(context.Context, CreateUserParams, Session) (user.User, error)
	FindUserByEmail(context.Context, string) (user.User, error)
	FindUserByID(context.Context, string) (user.User, error)
	CreateSession(context.Context, Session) error
	RotateSession(context.Context, []byte, Session, time.Time) (string, error)
	RevokeSession(context.Context, []byte, time.Time) error
}
