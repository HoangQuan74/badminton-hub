package authgrpc

import (
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func adminStaffResponse(value model.Staff) *identityV1.StaffResponse {
	return &identityV1.StaffResponse{
		Id:       value.ID,
		Email:    value.Email,
		FullName: value.FullName,
		IsAdmin:  value.IsAdmin,
		RoleId:   value.RoleID,
		RoleName: value.RoleName,
		Status:   value.Status,
	}
}

func currentAdminStaffResponse(value model.Staff) *identityV1.StaffResponse {
	response := adminStaffResponse(value)
	response.CreatedAt = value.CreatedAt.UTC().Format(time.RFC3339Nano)
	return response
}

func adminTokens(tokens auth.TokenPair) *identityV1.TokenPair {
	return &identityV1.TokenPair{
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		TokenType:        "Bearer",
		AccessExpiresAt:  tokens.AccessExpiresAt.UTC().Format(time.RFC3339),
		RefreshExpiresAt: tokens.RefreshExpiresAt.UTC().Format(time.RFC3339),
	}
}
