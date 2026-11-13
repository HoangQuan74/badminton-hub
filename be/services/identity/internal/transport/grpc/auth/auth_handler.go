package authgrpc

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

// sessionHandler shares protocol handling; each server supplies its own admission policy.
type sessionHandler struct {
	service *auth.Service
}

func (s *sessionHandler) login(ctx context.Context, request *identityV1.LoginRequest) (*identityV1.AuthResponse, error) {
	// Map the gRPC request to the domain input.
	input := auth.LoginInput{
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
	}

	// Execute login in the domain service.
	found, pair, err := s.service.Login(ctx, input)
	if err != nil {
		return nil, identityException.AuthGRPCMapper.Map(err)
	}

	// Map the domain result to the gRPC response.
	return authResponse(found, pair), nil
}

func (s *sessionHandler) refresh(ctx context.Context, request *identityV1.RefreshRequest) (*identityV1.RefreshResponse, error) {
	// Rotate the refresh session in the domain service.
	pair, err := s.service.Refresh(ctx, request.GetRefreshToken())
	if err != nil {
		return nil, identityException.AuthGRPCMapper.Map(err)
	}

	return &identityV1.RefreshResponse{
		Tokens: tokenPair(pair),
	}, nil
}

func (s *sessionHandler) logout(ctx context.Context, request *identityV1.LogoutRequest) (*identityV1.LogoutResponse, error) {
	// Revoke the refresh session in the domain service.
	if err := s.service.Logout(ctx, request.GetRefreshToken()); err != nil {
		return nil, identityException.AuthGRPCMapper.Map(err)
	}

	return &identityV1.LogoutResponse{}, nil
}

func (s *sessionHandler) currentUser(ctx context.Context, request *identityV1.GetCurrentUserRequest) (*identityV1.UserResponse, error) {
	// Resolve the account through the configured portal policy.
	found, err := s.service.CurrentPrincipal(ctx, request.GetAccessToken())
	if err != nil {
		return nil, identityException.AuthGRPCMapper.Map(err)
	}

	return userResponse(found), nil
}
