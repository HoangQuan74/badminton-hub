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
		CreatedAt: value.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
	}
}

func tokenPair(pair auth.TokenPair) *identityV1.TokenPair {
	// Map domain tokens to the gRPC token response.
	return &identityV1.TokenPair{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        "Bearer",
		AccessExpiresAt:  pair.AccessExpiresAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		RefreshExpiresAt: pair.RefreshExpiresAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
	}
}

func authResponse(value model.User, pair auth.TokenPair) *identityV1.AuthResponse {
	// Compose the authenticated user and token response.
	return &identityV1.AuthResponse{
		User:   userResponse(value),
		Tokens: tokenPair(pair),
	}
}
