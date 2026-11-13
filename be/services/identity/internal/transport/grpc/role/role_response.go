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

func roleListResponse(roles []model.Role) *identityV1.RoleListResponse {
	response := &identityV1.RoleListResponse{}
	for _, role := range roles {
		response.Roles = append(response.Roles, roleResponse(role))
	}

	return response
}
