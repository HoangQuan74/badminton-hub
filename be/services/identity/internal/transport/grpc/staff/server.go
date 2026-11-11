package staffgrpc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	staffDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	staffPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/staff"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type StaffServer struct {
	identityV1.UnimplementedIdentityStaffServiceServer
	service    *staffDomain.Service
	authorizer *identityGRPC.Authorizer
}

func NewStaffServer(cfg identityConfig.Config, pool *pgxpool.Pool) *StaffServer {
	// Build the staff service and shared admin authorizer.
	repository := staffPostgres.NewRepository(pool)
	return &StaffServer{
		service:    staffDomain.NewService(repository),
		authorizer: identityGRPC.NewAuthorizer(cfg, pool),
	}
}

func (s *StaffServer) authorizeToken(ctx context.Context, accessToken string) error {
	// Enforce admin access for staff management operations.
	_, err := s.authorizer.RequireAdmin(ctx, accessToken)
	return err
}
