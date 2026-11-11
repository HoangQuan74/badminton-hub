package role

import (
	"context"
	"errors"
	"strings"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

var (
	ErrInvalidInput = errors.New("role.invalid_input")
	ErrNotFound     = errors.New("role.not_found")
)

type Repository interface {
	ListRoles(context.Context) ([]model.Role, error)
	CreateRole(context.Context, string, string) (model.Role, error)
	GetRole(context.Context, string) (model.Role, error)
	UpdateRole(context.Context, string, string, string) (model.Role, error)
	DeleteRole(context.Context, string) error
	ReplaceRolePermissions(context.Context, string, []string) (model.Role, error)
	ListPermissions(context.Context) ([]model.Permission, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) ListRoles(ctx context.Context) ([]model.Role, error) {
	return s.repository.ListRoles(ctx)
}
func (s *Service) GetRole(ctx context.Context, id string) (model.Role, error) {
	return s.repository.GetRole(ctx, id)
}
func (s *Service) CreateRole(ctx context.Context, name, description string) (model.Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Role{}, ErrInvalidInput
	}
	return s.repository.CreateRole(ctx, name, strings.TrimSpace(description))
}
func (s *Service) UpdateRole(ctx context.Context, id, name, description string) (model.Role, error) {
	return s.repository.UpdateRole(ctx, id, strings.TrimSpace(name), strings.TrimSpace(description))
}
func (s *Service) DeleteRole(ctx context.Context, id string) error {
	return s.repository.DeleteRole(ctx, id)
}
func (s *Service) ReplaceRolePermissions(ctx context.Context, id string, permissions []string) (model.Role, error) {
	return s.repository.ReplaceRolePermissions(ctx, id, permissions)
}
func (s *Service) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	return s.repository.ListPermissions(ctx)
}
