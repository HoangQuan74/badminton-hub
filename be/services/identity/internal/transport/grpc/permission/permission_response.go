package permissiongrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func permissionResponse(permission model.Permission) *identityV1.Permission {
	return &identityV1.Permission{
		Id:          permission.ID,
		Module:      permission.Module,
		Path:        permission.Path,
		Action:      permission.Action,
		Description: permission.Description,
	}
}

func permissionListResponse(items []model.Permission) *identityV1.PermissionListResponse {
	response := &identityV1.PermissionListResponse{Permissions: make([]*identityV1.Permission, 0, len(items))}
	for _, item := range items {
		response.Permissions = append(response.Permissions, permissionResponse(item))
	}

	return response
}
