package permission

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type PermissionRepository interface {
	ListPermissions(ctx context.Context) ([]model.Permission, error)
	CreatePermission(ctx context.Context, input Input) (model.Permission, error)
	GetPermission(ctx context.Context, permissionID string) (model.Permission, error)
	UpdatePermission(ctx context.Context, permissionID string, input Input) (model.Permission, error)
	DeletePermission(ctx context.Context, permissionID string) error
}
