package authgrpc

import (
	"context"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *TenantServer) Login(ctx context.Context, request *identityV1.LoginRequest) (*identityV1.AuthResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(ctx context.Context, request *identityV1.LoginRequest) (*identityV1.AuthResponse, error) {
			// Authenticate only tenant operators.
			user, tokens, err := s.service.Login(ctx, auth.LoginInput{
				Email:    request.GetEmail(),
				Password: request.GetPassword(),
			})
			if err != nil {
				return nil, err
			}
			return authResponse(user, tokens), nil
		},
	)
}

func (s *TenantServer) Refresh(ctx context.Context, request *identityV1.RefreshRequest) (*identityV1.RefreshResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(ctx context.Context, request *identityV1.RefreshRequest) (*identityV1.RefreshResponse, error) {
			// Rotate the tenant session.
			tokens, err := s.service.Refresh(ctx, request.GetRefreshToken())
			if err != nil {
				return nil, err
			}
			return &identityV1.RefreshResponse{Tokens: tokenPair(tokens)}, nil
		},
	)
}

func (s *TenantServer) Logout(ctx context.Context, request *identityV1.LogoutRequest) (*identityV1.LogoutResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(ctx context.Context, request *identityV1.LogoutRequest) (*identityV1.LogoutResponse, error) {
			// Revoke the tenant session.
			if err := s.service.Logout(ctx, request.GetRefreshToken()); err != nil {
				return nil, err
			}
			return &identityV1.LogoutResponse{}, nil
		},
	)
}

func (s *TenantServer) GetCurrentUser(ctx context.Context, request *identityV1.GetCurrentUserRequest) (*identityV1.UserResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(ctx context.Context, request *identityV1.GetCurrentUserRequest) (*identityV1.UserResponse, error) {
			// Resolve the authenticated tenant operator.
			user, err := s.service.Me(ctx, request.GetAccessToken())
			if err != nil {
				return nil, err
			}
			return userResponse(user), nil
		},
	)
}
