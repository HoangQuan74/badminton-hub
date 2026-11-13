package usergrpc

import (
	"context"

	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	grpcHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/helper"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *UserServer) ListUsers(ctx context.Context, request *identityV1.ListUsersRequest) (*identityV1.UserListResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.UserGRPCMapper.Map(err)
	}

	// Build the shared pagination and filter query.
	query, err := grpcHelper.ParseListQuery(request)
	if err != nil {
		return nil, identityException.UserGRPCMapper.Map(identityException.ErrUserInvalidInput)
	}

	// Load the requested users from the domain service.
	items, err := s.service.ListUsers(ctx, query)
	if err != nil {
		return nil, identityException.UserGRPCMapper.Map(err)
	}

	return userListResponse(items), nil
}

func (s *UserServer) CreateUser(ctx context.Context, request *identityV1.CreateUserRequest) (*identityV1.UserResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.UserGRPCMapper.Map(err)
	}

	// Map the gRPC request to the domain input.
	input := userDomain.CreateInput{
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
		FullName: request.GetFullName(),
		Phone:    request.GetPhone(),
	}

	// Create the user in the domain service.
	item, err := s.service.CreateUser(ctx, input)
	if err != nil {
		return nil, identityException.UserGRPCMapper.Map(err)
	}

	return grpcHelper.MapUserResponse(item), nil
}

func (s *UserServer) GetUser(ctx context.Context, request *identityV1.GetUserRequest) (*identityV1.UserResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.UserGRPCMapper.Map(err)
	}

	// Load the requested user from the domain service.
	item, err := s.service.GetUser(ctx, request.GetUserId())
	if err != nil {
		return nil, identityException.UserGRPCMapper.Map(err)
	}

	return grpcHelper.MapUserResponse(item), nil
}

func (s *UserServer) UpdateUser(ctx context.Context, request *identityV1.UpdateUserRequest) (*identityV1.UserResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.UserGRPCMapper.Map(err)
	}

	// Map the gRPC request to the domain input.
	input := userDomain.UpdateInput{
		ID:       request.GetUserId(),
		Email:    request.GetEmail(),
		FullName: request.GetFullName(),
		Phone:    request.GetPhone(),
		Status:   request.GetStatus(),
	}

	// Update the user in the domain service.
	item, err := s.service.UpdateUser(ctx, input)
	if err != nil {
		return nil, identityException.UserGRPCMapper.Map(err)
	}

	return grpcHelper.MapUserResponse(item), nil
}

func (s *UserServer) DeleteUser(ctx context.Context, request *identityV1.DeleteUserRequest) (*identityV1.Empty, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.UserGRPCMapper.Map(err)
	}

	// Delete the user in the domain service.
	if err := s.service.DeleteUser(ctx, request.GetUserId()); err != nil {
		return nil, identityException.UserGRPCMapper.Map(err)
	}

	return &identityV1.Empty{}, nil
}
