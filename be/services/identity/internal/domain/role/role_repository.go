package role

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type RoleRepository interface {
	ListRoles(ctx context.Context) ([]model.Role, error)
	CreateRole(ctx context.Context, name, description string) (model.Role, error)
	GetRole(ctx context.Context, roleID string) (model.Role, error)
	UpdateRole(ctx context.Context, roleID, name, description string) (model.Role, error)
	DeleteRole(ctx context.Context, roleID string) error
	ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []string) (model.Role, error)
}
