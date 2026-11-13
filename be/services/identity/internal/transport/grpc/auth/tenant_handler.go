package authgrpc

import (
	"context"

	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *TenantServer) Login(ctx context.Context, request *identityV1.LoginRequest) (*identityV1.AuthResponse, error) {
	return s.session.login(ctx, request)
}

func (s *TenantServer) Refresh(ctx context.Context, request *identityV1.RefreshRequest) (*identityV1.RefreshResponse, error) {
	return s.session.refresh(ctx, request)
}

func (s *TenantServer) Logout(ctx context.Context, request *identityV1.LogoutRequest) (*identityV1.LogoutResponse, error) {
	return s.session.logout(ctx, request)
}

func (s *TenantServer) GetCurrentUser(ctx context.Context, request *identityV1.GetCurrentUserRequest) (*identityV1.UserResponse, error) {
	return s.session.currentUser(ctx, request)
}
