package rolegrpc

import (
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func roleResponse(role model.Role) *identityV1.Role {
	return &identityV1.Role{
		Id:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func permissionResponse(permission model.Permission) *identityV1.Permission {
	return &identityV1.Permission{
		Id:          permission.ID,
		Module:      permission.Module,
		Path:        permission.Path,
		Action:      permission.Action,
		Description: permission.Description,
	}
}
