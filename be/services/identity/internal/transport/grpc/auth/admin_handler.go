package authgrpc

import (
	"context"

	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *AdminServer) Login(ctx context.Context, request *identityV1.AdminLoginRequest) (*identityV1.AdminAuthResponse, error) {
	// Authenticate the admin account.
	staff, tokens, err := s.service.Login(ctx, request.GetEmail(), request.GetPassword())
	if err != nil {
		return nil, identityException.AdminGRPCMapper.Map(err)
	}

	return adminAuthResponse(staff, tokens), nil
}

func (s *AdminServer) Refresh(ctx context.Context, request *identityV1.AdminRefreshRequest) (*identityV1.AdminRefreshResponse, error) {
	tokens, err := s.service.Refresh(ctx, request.GetRefreshToken())
	if err != nil {
		return nil, identityException.AdminGRPCMapper.Map(err)
	}

	return &identityV1.AdminRefreshResponse{Tokens: adminTokens(tokens)}, nil
}

func (s *AdminServer) Logout(ctx context.Context, request *identityV1.AdminLogoutRequest) (*identityV1.AdminLogoutResponse, error) {
	if err := s.service.Logout(ctx, request.GetRefreshToken()); err != nil {
		return nil, identityException.AdminGRPCMapper.Map(err)
	}

	return &identityV1.AdminLogoutResponse{}, nil
}

func (s *AdminServer) GetCurrentStaff(ctx context.Context, request *identityV1.CurrentStaffRequest) (*identityV1.StaffResponse, error) {
	// Verify the authenticated actor.
	item, err := s.authorizer.RequireAdmin(ctx, request.GetAccessToken())
	if err != nil {
		return nil, identityException.AdminGRPCMapper.Map(err)
	}

	return currentAdminStaffResponse(item), nil
}

func (s *AdminServer) ChangePassword(ctx context.Context, request *identityV1.ChangePasswordRequest) (*identityV1.Empty, error) {
	// Verify the authenticated actor.
	item, err := s.authorizer.RequireAdmin(ctx, request.GetAccessToken())
	if err != nil {
		return nil, identityException.AdminGRPCMapper.Map(err)
	}

	// Change password.
	if err := s.service.ChangePassword(ctx, item.ID, request.GetCurrentPassword(), request.GetNewPassword()); err != nil {
		return nil, identityException.AdminGRPCMapper.Map(err)
	}

	return &identityV1.Empty{}, nil
}
