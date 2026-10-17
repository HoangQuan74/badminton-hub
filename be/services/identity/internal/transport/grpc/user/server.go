package usergrpc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	userPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/user"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type UserServer struct {
	identityV1.UnimplementedIdentityUserServiceServer
	service    *userDomain.Service
	authorizer *identityGRPC.Authorizer
}

func NewUserServer(cfg identityConfig.Config, pool *pgxpool.Pool) *UserServer {
	// Build the user service and shared admin authorizer.
	return &UserServer{
		service:    userDomain.NewService(userPostgres.NewRepository(pool)),
		authorizer: identityGRPC.NewAuthorizer(cfg, pool),
	}
}

func (s *UserServer) ListUsers(ctx context.Context, request *identityV1.ListUsersRequest) (*identityV1.UserListResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.ListUsersRequest](s.authorizeToken),
		func(ctx context.Context, request *identityV1.ListUsersRequest) (*identityV1.UserListResponse, error) {
			// Build the shared pagination and filter query.
			query, err := model.NewListQuery(
				int(request.GetLimit()),
				int(request.GetOffset()),
				request.GetSearch(),
				request.GetCreatedFrom(),
				request.GetCreatedTo(),
			)
			if err != nil {
				return nil, userDomain.ErrInvalidInput
			}

			// Load the requested users from the domain service.
			items, err := s.service.List(ctx, query)
			if err != nil {
				return nil, err
			}

			// Map the users to the gRPC list response.
			return userListResponse(items), nil
		},
	)
}

func (s *UserServer) CreateUser(ctx context.Context, request *identityV1.CreateUserRequest) (*identityV1.UserResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.CreateUserRequest](s.authorizeToken),
		func(ctx context.Context, request *identityV1.CreateUserRequest) (*identityV1.UserResponse, error) {
			// Map the gRPC request to the domain input.
			input := userDomain.CreateInput{
				Email:    request.GetEmail(),
				Password: request.GetPassword(),
				FullName: request.GetFullName(),
				Phone:    request.GetPhone(),
			}

			// Create the user in the domain service.
			item, err := s.service.Create(ctx, input)
			if err != nil {
				return nil, err
			}

			// Map the created user to the gRPC response.
			return userResponse(item), nil
		},
	)
}

func (s *UserServer) GetUser(ctx context.Context, request *identityV1.GetUserRequest) (*identityV1.UserResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.GetUserRequest](s.authorizeToken),
		func(ctx context.Context, request *identityV1.GetUserRequest) (*identityV1.UserResponse, error) {
			// Load the requested user from the domain service.
			item, err := s.service.Get(ctx, request.GetUserId())
			if err != nil {
				return nil, err
			}

			// Map the user to the gRPC response.
			return userResponse(item), nil
		},
	)
}

func (s *UserServer) UpdateUser(ctx context.Context, request *identityV1.UpdateUserRequest) (*identityV1.UserResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.UpdateUserRequest](s.authorizeToken),
		func(ctx context.Context, request *identityV1.UpdateUserRequest) (*identityV1.UserResponse, error) {
			// Map the gRPC request to the domain input.
			input := userDomain.UpdateInput{
				ID:       request.GetUserId(),
				Email:    request.GetEmail(),
				FullName: request.GetFullName(),
				Phone:    request.GetPhone(),
				Status:   request.GetStatus(),
			}

			// Update the user in the domain service.
			item, err := s.service.Update(ctx, input)
			if err != nil {
				return nil, err
			}

			// Map the updated user to the gRPC response.
			return userResponse(item), nil
		},
	)
}

func (s *UserServer) DeleteUser(ctx context.Context, request *identityV1.DeleteUserRequest) (*identityV1.Empty, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.DeleteUserRequest](s.authorizeToken),
		func(ctx context.Context, request *identityV1.DeleteUserRequest) (*identityV1.Empty, error) {
			// Delete the user in the domain service.
			if err := s.service.Delete(ctx, request.GetUserId()); err != nil {
				return nil, err
			}

			// Return the empty deletion acknowledgment.
			return &identityV1.Empty{}, nil
		},
	)
}

func (s *UserServer) authorizeToken(ctx context.Context, accessToken string) error {
	// Enforce admin access for user management operations.
	_, err := s.authorizer.RequireAdmin(ctx, accessToken)
	return err
}
