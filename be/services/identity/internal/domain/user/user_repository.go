package user

import (
	"context"

	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type CreateParams struct {
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	Status       string
}

type UserRepository interface {
	ListUsers(ctx context.Context, query model.ListQuery) ([]model.User, error)
	GetUser(ctx context.Context, userID string) (model.User, error)
	CreateUser(ctx context.Context, params CreateParams) (model.User, error)
	UpdateUser(ctx context.Context, input UpdateInput) (model.User, error)
	DeleteUser(ctx context.Context, userID string) error
}

// RegistrationRepository persists a new account and its initial session atomically.
type RegistrationRepository interface {
	CreateUserWithSession(ctx context.Context, params RegistrationParams, session model.Session) (model.User, error)
}
type RegistrationParams struct {
	ID           string
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
}
