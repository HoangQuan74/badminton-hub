package rolegrpc

import (
	"context"

	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *RoleServer) ListRoles(ctx context.Context, request *identityV1.ListRolesRequest) (*identityV1.RoleListResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	roles, err := s.service.ListRoles(ctx)
	if err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	return roleListResponse(roles), nil
}

func (s *RoleServer) CreateRole(ctx context.Context, request *identityV1.CreateRoleRequest) (*identityV1.RoleResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	role, err := s.service.CreateRole(ctx, request.GetName(), request.GetDescription())
	if err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	return &identityV1.RoleResponse{Role: roleResponse(role)}, nil
}

func (s *RoleServer) GetRole(ctx context.Context, request *identityV1.GetRoleRequest) (*identityV1.RoleResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	role, err := s.service.GetRole(ctx, request.GetRoleId())
	if err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	return &identityV1.RoleResponse{Role: roleResponse(role)}, nil
}

func (s *RoleServer) UpdateRole(ctx context.Context, request *identityV1.UpdateRoleRequest) (*identityV1.RoleResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	role, err := s.service.UpdateRole(ctx, request.GetRoleId(), request.GetName(), request.GetDescription())
	if err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	return &identityV1.RoleResponse{Role: roleResponse(role)}, nil
}

func (s *RoleServer) DeleteRole(ctx context.Context, request *identityV1.DeleteRoleRequest) (*identityV1.Empty, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	if err := s.service.DeleteRole(ctx, request.GetRoleId()); err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	return &identityV1.Empty{}, nil
}

func (s *RoleServer) ReplaceRolePermissions(
	ctx context.Context,
	request *identityV1.ReplaceRolePermissionsRequest,
) (*identityV1.RoleResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	role, err := s.service.ReplaceRolePermissions(ctx, request.GetRoleId(), request.GetPermissionIds())
	if err != nil {
		return nil, identityException.RoleGRPCMapper.Map(err)
	}

	return &identityV1.RoleResponse{Role: roleResponse(role)}, nil
}
