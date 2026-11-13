package authgrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	grpcHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/helper"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func userResponse(value model.User) *identityV1.UserResponse {
	response := grpcHelper.MapUserResponse(value)
	response.Roles = roles(value)
	response.IsAdmin = value.IsAdmin
	return response
}

func roles(value model.User) []string {
	if value.RoleName == "" {
		return []string{constant.RoleCustomer}
	}

	return []string{value.RoleName}
}

func authResponse(value model.User, pair model.TokenPair) *identityV1.AuthResponse {
	// Compose the authenticated user and token response.
	return &identityV1.AuthResponse{
		User:   userResponse(value),
		Tokens: tokenPair(pair),
	}
}
