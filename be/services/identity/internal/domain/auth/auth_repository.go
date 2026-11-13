package auth

import (
	"context"
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type AuthRepository interface {
	FindCredentialsByEmail(ctx context.Context, email string) (model.User, error)
	GetPrincipal(ctx context.Context, principalID string) (model.User, error)
	CreateSession(ctx context.Context, session model.Session) error
	WithinSessionTransaction(ctx context.Context, execute func(AuthSessionTransaction) error) error
	RevokeSession(ctx context.Context, tokenHash []byte, now time.Time) error
}

type AuthSessionTransaction interface {
	LockSessionOwner(ctx context.Context, tokenHash []byte, now time.Time) (string, error)
	GetPrincipal(ctx context.Context, principalID string) (model.User, error)
	RevokeSession(ctx context.Context, tokenHash []byte, now time.Time) error
	CreateSession(ctx context.Context, session model.Session) error
}
