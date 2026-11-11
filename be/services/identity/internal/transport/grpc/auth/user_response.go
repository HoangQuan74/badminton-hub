package authgrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func userResponse(value model.User) *identityV1.UserResponse {
	// Map the domain user to the public user response.
	return &identityV1.UserResponse{
		Id:        value.ID,
		Email:     value.Email,
		FullName:  value.FullName,
		Phone:     value.Phone,
		Status:    string(value.Status),
		Roles:     roles(value),
		IsAdmin:   value.IsAdmin,
		CreatedAt: value.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
	}
}

func roles(value model.User) []string {
	if value.RoleName == "" {
		return []string{"customer"}
	}
	return []string{value.RoleName}
}

func authResponse(value model.User, pair auth.TokenPair) *identityV1.AuthResponse {
	// Compose the authenticated user and token response.
	return &identityV1.AuthResponse{
		User:   userResponse(value),
		Tokens: tokenPair(pair),
	}
}
