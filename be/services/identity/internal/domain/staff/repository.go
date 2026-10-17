package staff

import (
	"context"

	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type CreateParams struct {
	Email, PasswordHash, FullName, RoleID string
	IsAdmin                               bool
}
type Repository interface {
	List(context.Context, model.ListQuery) ([]model.Staff, error)
	FindByID(context.Context, string) (model.Staff, error)
	Create(context.Context, CreateParams) (model.Staff, error)
	Update(context.Context, UpdateInput) (model.Staff, error)
	UpdateAccess(context.Context, AccessInput) (model.Staff, error)
	Delete(context.Context, string) error
}
