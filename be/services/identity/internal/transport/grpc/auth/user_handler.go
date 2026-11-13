package authgrpc

import (
	"context"

	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *UserServer) Register(ctx context.Context, request *identityV1.RegisterRequest) (*identityV1.AuthResponse, error) {
	// Map the gRPC request to the domain input.
	input := userDomain.RegisterInput{
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
		FullName: request.GetFullName(),
		Phone:    request.GetPhone(),
	}

	// Execute registration in the domain service.
	created, pair, err := s.registration.RegisterUser(ctx, input)
	if err != nil {
		return nil, identityException.AuthGRPCMapper.Map(err)
	}

	// Map the domain result to the gRPC response.
	return authResponse(created, pair), nil
}

func (s *UserServer) Login(ctx context.Context, request *identityV1.LoginRequest) (*identityV1.AuthResponse, error) {
	return s.session.login(ctx, request)
}

func (s *UserServer) Refresh(ctx context.Context, request *identityV1.RefreshRequest) (*identityV1.RefreshResponse, error) {
	return s.session.refresh(ctx, request)
}

func (s *UserServer) Logout(ctx context.Context, request *identityV1.LogoutRequest) (*identityV1.LogoutResponse, error) {
	return s.session.logout(ctx, request)
}

func (s *UserServer) GetCurrentUser(ctx context.Context, request *identityV1.GetCurrentUserRequest) (*identityV1.UserResponse, error) {
	return s.session.currentUser(ctx, request)
}
