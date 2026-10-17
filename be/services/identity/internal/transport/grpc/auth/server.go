package authgrpc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	authPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/auth"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type AuthServer struct {
	identityV1.UnimplementedIdentityAuthServiceServer
	service *auth.Service
}

func NewAuthServer(cfg identityConfig.Config, pool *pgxpool.Pool) *AuthServer {
	// Build the auth repository and token manager dependencies.
	repository := authPostgres.NewAuthRepository(pool)
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	return &AuthServer{
		service: auth.NewService(repository, tokens),
	}
}

func (s *AuthServer) Register(
	ctx context.Context,
	request *identityV1.RegisterRequest,
) (*identityV1.AuthResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(ctx context.Context, request *identityV1.RegisterRequest) (*identityV1.AuthResponse, error) {
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

func (s *AuthServer) Login(
	ctx context.Context,
	request *identityV1.LoginRequest,
) (*identityV1.AuthResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(ctx context.Context, request *identityV1.LoginRequest) (*identityV1.AuthResponse, error) {
			// Map the gRPC request to the domain input.
			input := auth.LoginInput{
				Email:    request.GetEmail(),
				Password: request.GetPassword(),
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

func (s *AuthServer) Refresh(
	ctx context.Context,
	request *identityV1.RefreshRequest,
) (*identityV1.RefreshResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(ctx context.Context, request *identityV1.RefreshRequest) (*identityV1.RefreshResponse, error) {
			// Rotate the refresh session in the domain service.
			pair, err := s.service.Refresh(ctx, request.GetRefreshToken())
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

func (s *AuthServer) Logout(
	ctx context.Context,
	request *identityV1.LogoutRequest,
) (*identityV1.LogoutResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(ctx context.Context, request *identityV1.LogoutRequest) (*identityV1.LogoutResponse, error) {
			// Revoke the refresh session in the domain service.
			if err := s.service.Logout(ctx, request.GetRefreshToken()); err != nil {
				return nil, err
			}

			// Return the empty logout acknowledgment.
			return &identityV1.LogoutResponse{}, nil
		},
	)
}

func (s *AuthServer) GetCurrentUser(
	ctx context.Context,
	request *identityV1.GetCurrentUserRequest,
) (*identityV1.UserResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(ctx context.Context, request *identityV1.GetCurrentUserRequest) (*identityV1.UserResponse, error) {
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

			// Map the user to the gRPC response.
			return userResponse(found), nil
		},
	)
}
