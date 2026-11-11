package rolegrpc

import (
	"context"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *Handler) ListRoles(
	ctx context.Context,
	request *identityV1.ListRolesRequest,
) (*identityV1.RoleListResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.ListRolesRequest](s.authorizeToken),
		func(
			ctx context.Context,
			request *identityV1.ListRolesRequest,
		) (*identityV1.RoleListResponse, error) {
			// Load roles.
			roles, err := s.service.ListRoles(ctx)
			if err != nil {
				return nil, err
			}
			response := &identityV1.RoleListResponse{}
			for _, role := range roles {
				response.Roles = append(response.Roles, roleResponse(role))
			}
			return response, nil
		},
	)
}

func (s *Handler) CreateRole(
	ctx context.Context,
	request *identityV1.CreateRoleRequest,
) (*identityV1.RoleResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.CreateRoleRequest](s.authorizeToken),
		func(
			ctx context.Context,
			request *identityV1.CreateRoleRequest,
		) (*identityV1.RoleResponse, error) {
			// Create role.
			role, err := s.service.CreateRole(
				ctx,
				request.GetName(),
				request.GetDescription(),
			)
			if err != nil {
				return nil, err
			}
			return &identityV1.RoleResponse{Role: roleResponse(role)}, nil
		},
	)
}

func (s *Handler) GetRole(
	ctx context.Context,
	request *identityV1.GetRoleRequest,
) (*identityV1.RoleResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.GetRoleRequest](s.authorizeToken),
		func(
			ctx context.Context,
			request *identityV1.GetRoleRequest,
		) (*identityV1.RoleResponse, error) {
			// Load role.
			role, err := s.service.GetRole(ctx, request.GetRoleId())
			if err != nil {
				return nil, err
			}
			return &identityV1.RoleResponse{Role: roleResponse(role)}, nil
		},
	)
}

func (s *Handler) UpdateRole(
	ctx context.Context,
	request *identityV1.UpdateRoleRequest,
) (*identityV1.RoleResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.UpdateRoleRequest](s.authorizeToken),
		func(
			ctx context.Context,
			request *identityV1.UpdateRoleRequest,
		) (*identityV1.RoleResponse, error) {
			// Update role.
			role, err := s.service.UpdateRole(
				ctx,
				request.GetRoleId(),
				request.GetName(),
				request.GetDescription(),
			)
			if err != nil {
				return nil, err
			}
			return &identityV1.RoleResponse{Role: roleResponse(role)}, nil
		},
	)
}

func (s *Handler) DeleteRole(
	ctx context.Context,
	request *identityV1.DeleteRoleRequest,
) (*identityV1.Empty, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.DeleteRoleRequest](s.authorizeToken),
		func(
			ctx context.Context,
			request *identityV1.DeleteRoleRequest,
		) (*identityV1.Empty, error) {
			// Delete role.
			if err := s.service.DeleteRole(ctx, request.GetRoleId()); err != nil {
				return nil, err
			}
			return &identityV1.Empty{}, nil
		},
	)
}

func (s *Handler) ReplaceRolePermissions(
	ctx context.Context,
	request *identityV1.ReplaceRolePermissionsRequest,
) (*identityV1.RoleResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.ReplaceRolePermissionsRequest](s.authorizeToken),
		func(
			ctx context.Context,
			request *identityV1.ReplaceRolePermissionsRequest,
		) (*identityV1.RoleResponse, error) {
			// Replace role permissions.
			role, err := s.service.ReplaceRolePermissions(
				ctx,
				request.GetRoleId(),
				request.GetPermissionIds(),
			)
			if err != nil {
				return nil, err
			}
			return &identityV1.RoleResponse{Role: roleResponse(role)}, nil
		},
	)
}

func (s *Handler) ListPermissions(
	ctx context.Context,
	request *identityV1.ListPermissionsRequest,
) (*identityV1.PermissionListResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.ListPermissionsRequest](s.authorizeToken),
		func(
			ctx context.Context,
			request *identityV1.ListPermissionsRequest,
		) (*identityV1.PermissionListResponse, error) {
			// Load permissions.
			permissions, err := s.service.ListPermissions(ctx)
			if err != nil {
				return nil, err
			}
			response := &identityV1.PermissionListResponse{}
			for _, permission := range permissions {
				response.Permissions = append(response.Permissions, permissionResponse(permission))
			}
			return response, nil
		},
	)
}
