package role

import (
	"context"
	"strings"

	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type Service struct{ repository RoleRepository }

func NewService(repository RoleRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListRoles(ctx context.Context) ([]model.Role, error) {
	return s.repository.ListRoles(ctx)
}

func (s *Service) GetRole(ctx context.Context, id string) (model.Role, error) {
	return s.repository.GetRole(ctx, id)
}

func (s *Service) CreateRole(ctx context.Context, name, description string) (model.Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Role{}, identityException.ErrRoleInvalidInput
	}

	return s.repository.CreateRole(
		ctx,
		name,
		strings.TrimSpace(description),
	)
}

func (s *Service) UpdateRole(ctx context.Context, id, name, description string) (model.Role, error) {
	return s.repository.UpdateRole(
		ctx,
		id,
		strings.TrimSpace(name),
		strings.TrimSpace(description),
	)
}

func (s *Service) DeleteRole(ctx context.Context, id string) error {
	return s.repository.DeleteRole(ctx, id)
}

func (s *Service) ReplaceRolePermissions(ctx context.Context, id string, permissions []string) (model.Role, error) {
	return s.repository.ReplaceRolePermissions(
		ctx,
		id,
		permissions,
	)
}
