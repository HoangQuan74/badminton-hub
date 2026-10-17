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

type Repository interface {
	List(context.Context, model.ListQuery) ([]model.User, error)
	FindByID(context.Context, string) (model.User, error)
	Create(context.Context, CreateParams) (model.User, error)
	Update(context.Context, UpdateInput) (model.User, error)
	Delete(context.Context, string) error
}
