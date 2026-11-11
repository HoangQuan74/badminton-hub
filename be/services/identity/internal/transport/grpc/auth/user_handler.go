package authgrpc

import (
	"context"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *UserServer) Register(
	ctx context.Context,
	request *identityV1.RegisterRequest,
) (*identityV1.AuthResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *identityV1.RegisterRequest,
		) (*identityV1.AuthResponse, error) {
			// Map the gRPC request to the domain input.
			input := auth.RegisterInput{
				Email:    request.GetEmail(),
				Password: request.GetPassword(),
				FullName: request.GetFullName(),
				Phone:    request.GetPhone(),
			}

			// Execute registration in the domain service.
			created, pair, err := s.service.Register(ctx, input)
			if err != nil {
				return nil, err
			}

			// Map the domain result to the gRPC response.
			return authResponse(created, pair), nil
		},
	)
}

func (s *UserServer) Login(
	ctx context.Context,
	request *identityV1.LoginRequest,
) (*identityV1.AuthResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *identityV1.LoginRequest,
		) (*identityV1.AuthResponse, error) {
			// Map the gRPC request to the domain input.
			input := auth.LoginInput{
				Email:    request.GetEmail(),
				Password: request.GetPassword(),
				Portal:   request.GetPortal(),
			}

			// Execute login in the domain service.
			found, pair, err := s.service.Login(ctx, input)
			if err != nil {
				return nil, err
			}

			// Map the domain result to the gRPC response.
			return authResponse(found, pair), nil
		},
	)
}

func (s *UserServer) Refresh(
	ctx context.Context,
	request *identityV1.RefreshRequest,
) (*identityV1.RefreshResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *identityV1.RefreshRequest,
		) (*identityV1.RefreshResponse, error) {
			// Rotate the refresh session in the domain service.
			pair, err := s.service.Refresh(
				ctx,
				request.GetRefreshToken(),
				request.GetPortal(),
			)
			if err != nil {
				return nil, err
			}
			// Map the replacement tokens to the gRPC response.
			return &identityV1.RefreshResponse{
				Tokens: tokenPair(pair),
			}, nil
		},
	)
}

func (s *UserServer) Logout(
	ctx context.Context,
	request *identityV1.LogoutRequest,
) (*identityV1.LogoutResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *identityV1.LogoutRequest,
		) (*identityV1.LogoutResponse, error) {
			// Revoke the refresh session in the domain service.
			if err := s.service.Logout(ctx, request.GetRefreshToken()); err != nil {
				return nil, err
			}

			// Return the empty logout acknowledgment.
			return &identityV1.LogoutResponse{}, nil
		},
	)
}

func (s *UserServer) GetCurrentUser(
	ctx context.Context,
	request *identityV1.GetCurrentUserRequest,
) (*identityV1.UserResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *identityV1.GetCurrentUserRequest,
		) (*identityV1.UserResponse, error) {
			// Parse and verify the access token.
			userID, err := s.service.ParseAccessToken(request.GetAccessToken())
			if err != nil {
				return nil, err
			}
			// Load the authenticated user.
			found, err := s.service.Me(ctx, userID)
			if err != nil {
				return nil, err
			}
			if !auth.IsAllowedPortal(found, request.GetPortal()) {
				return nil, auth.ErrWrongPortal
			}

			// Map the user to the gRPC response.
			return userResponse(found), nil
		},
	)
}
