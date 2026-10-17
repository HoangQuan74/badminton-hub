package staffgrpc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	staffDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	staffPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/staff"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type StaffServer struct {
	identityV1.UnimplementedIdentityAdminServiceServer
	service    *staffDomain.Service
	pool       *pgxpool.Pool
	authorizer *identityGRPC.Authorizer
	tokens     *auth.TokenManager
}

func NewStaffServer(cfg identityConfig.Config, pool *pgxpool.Pool) *StaffServer {
	// Build the staff service and shared admin authorizer.
	return &StaffServer{
		service:    staffDomain.NewService(staffPostgres.NewRepository(pool)),
		pool:       pool,
		authorizer: identityGRPC.NewAuthorizer(cfg, pool),
		tokens:     auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL),
	}
}

func (s *StaffServer) ListStaff(ctx context.Context, request *identityV1.ListStaffRequest) (*identityV1.StaffListResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.ListStaffRequest](s.authorizeToken),
		func(ctx context.Context, request *identityV1.ListStaffRequest) (*identityV1.StaffListResponse, error) {
			// Build the shared pagination and filter query.
			query, err := model.NewListQuery(
				int(request.GetLimit()),
				int(request.GetOffset()),
				request.GetSearch(),
				request.GetCreatedFrom(),
				request.GetCreatedTo(),
			)
			if err != nil {
				return nil, staffDomain.ErrInvalidInput
			}

			// Load the requested staff members from the domain service.
			items, err := s.service.List(ctx, query)
			if err != nil {
				return nil, err
			}

			// Map the staff members to the gRPC list response.
			return staffListResponse(items), nil
		},
	)
}

func (s *StaffServer) CreateStaff(ctx context.Context, request *identityV1.CreateStaffRequest) (*identityV1.StaffResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.CreateStaffRequest](s.authorizeToken),
		func(ctx context.Context, request *identityV1.CreateStaffRequest) (*identityV1.StaffResponse, error) {
			// Map the gRPC request to the domain input.
			input := staffDomain.CreateInput{
				Email:    request.GetEmail(),
				Password: request.GetPassword(),
				FullName: request.GetFullName(),
				IsAdmin:  request.GetIsAdmin(),
				RoleID:   request.GetRoleId(),
			}

			// Create the staff member in the domain service.
			item, err := s.service.Create(ctx, input)
			if err != nil {
				return nil, err
			}

			// Map the created staff member to the gRPC response.
			return staffResponse(item), nil
		},
	)
}

func (s *StaffServer) GetStaff(ctx context.Context, request *identityV1.GetStaffRequest) (*identityV1.StaffResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.GetStaffRequest](s.authorizeToken),
		func(ctx context.Context, request *identityV1.GetStaffRequest) (*identityV1.StaffResponse, error) {
			// Load the requested staff member from the domain service.
			item, err := s.service.Get(ctx, request.GetStaffId())
			if err != nil {
				return nil, err
			}

			// Map the staff member to the gRPC response.
			return staffResponse(item), nil
		},
	)
}

func (s *StaffServer) UpdateStaff(ctx context.Context, request *identityV1.UpdateStaffRequest) (*identityV1.StaffResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.UpdateStaffRequest](s.authorizeToken),
		func(ctx context.Context, request *identityV1.UpdateStaffRequest) (*identityV1.StaffResponse, error) {
			// Map the gRPC request to the domain input.
			input := staffDomain.UpdateInput{
				ID:       request.GetStaffId(),
				Email:    request.GetEmail(),
				FullName: request.GetFullName(),
				Status:   request.GetStatus(),
			}

			// Update the staff member in the domain service.
			item, err := s.service.Update(ctx, input)
			if err != nil {
				return nil, err
			}

			// Map the updated staff member to the gRPC response.
			return staffResponse(item), nil
		},
	)
}

func (s *StaffServer) DeleteStaff(ctx context.Context, request *identityV1.DeleteStaffRequest) (*identityV1.Empty, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.DeleteStaffRequest](s.authorizeToken),
		func(ctx context.Context, request *identityV1.DeleteStaffRequest) (*identityV1.Empty, error) {
			// Delete the staff member in the domain service.
			if err := s.service.Delete(ctx, request.GetStaffId()); err != nil {
				return nil, err
			}

			// Return the empty deletion acknowledgment.
			return &identityV1.Empty{}, nil
		},
	)
}

func (s *StaffServer) UpdateStaffAccess(ctx context.Context, request *identityV1.UpdateStaffAccessRequest) (*identityV1.StaffResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		grpcEndpoint.Authorize[*identityV1.UpdateStaffAccessRequest](s.authorizeToken),
		func(ctx context.Context, request *identityV1.UpdateStaffAccessRequest) (*identityV1.StaffResponse, error) {
			// Map the gRPC request to the domain input.
			input := staffDomain.AccessInput{
				ID:      request.GetStaffId(),
				IsAdmin: request.GetIsAdmin(),
				RoleID:  request.GetRoleId(),
			}

			// Update access in the domain service.
			item, err := s.service.UpdateAccess(ctx, input)
			if err != nil {
				return nil, err
			}

			// Map the updated staff member to the gRPC response.
			return staffResponse(item), nil
		},
	)
}

func (s *StaffServer) authorizeToken(ctx context.Context, accessToken string) error {
	// Enforce admin access for staff management operations.
	_, err := s.authorizer.RequireAdmin(ctx, accessToken)
	return err
}
