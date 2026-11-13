package permissiongrpc

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/permission"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *PermissionServer) ListPermissions(ctx context.Context, request *identityV1.ListPermissionsRequest) (*identityV1.PermissionListResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.PermissionGRPCMapper.Map(err)
	}

	items, err := s.service.ListPermissions(ctx)
	if err != nil {
		return nil, identityException.PermissionGRPCMapper.Map(err)
	}

	return permissionListResponse(items), nil
}

func (s *PermissionServer) CreatePermission(ctx context.Context, request *identityV1.CreatePermissionRequest) (*identityV1.PermissionResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.PermissionGRPCMapper.Map(err)
	}

	input := permission.Input{
		Module:      request.GetModule(),
		Path:        request.GetPath(),
		Action:      request.GetAction(),
		Description: request.GetDescription(),
	}

	item, err := s.service.CreatePermission(ctx, input)
	if err != nil {
		return nil, identityException.PermissionGRPCMapper.Map(err)
	}

	return &identityV1.PermissionResponse{Permission: permissionResponse(item)}, nil
}

func (s *PermissionServer) GetPermission(ctx context.Context, request *identityV1.GetPermissionRequest) (*identityV1.PermissionResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.PermissionGRPCMapper.Map(err)
	}

	item, err := s.service.GetPermission(ctx, request.GetPermissionId())
	if err != nil {
		return nil, identityException.PermissionGRPCMapper.Map(err)
	}

	return &identityV1.PermissionResponse{Permission: permissionResponse(item)}, nil
}

func (s *PermissionServer) UpdatePermission(ctx context.Context, request *identityV1.UpdatePermissionRequest) (*identityV1.PermissionResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.PermissionGRPCMapper.Map(err)
	}

	input := permission.Input{
		Module:      request.GetModule(),
		Path:        request.GetPath(),
		Action:      request.GetAction(),
		Description: request.GetDescription(),
	}

	item, err := s.service.UpdatePermission(ctx, request.GetPermissionId(), input)
	if err != nil {
		return nil, identityException.PermissionGRPCMapper.Map(err)
	}

	return &identityV1.PermissionResponse{Permission: permissionResponse(item)}, nil
}

func (s *PermissionServer) DeletePermission(ctx context.Context, request *identityV1.DeletePermissionRequest) (*identityV1.Empty, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.PermissionGRPCMapper.Map(err)
	}

	if err := s.service.DeletePermission(ctx, request.GetPermissionId()); err != nil {
		return nil, identityException.PermissionGRPCMapper.Map(err)
	}

	return &identityV1.Empty{}, nil
}
