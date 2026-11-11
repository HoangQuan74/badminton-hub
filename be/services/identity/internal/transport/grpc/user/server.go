package usergrpc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
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

func (s *UserServer) authorizeToken(ctx context.Context, accessToken string) error {
	// Enforce admin access for user management operations.
	_, err := s.authorizer.RequireAdmin(ctx, accessToken)
	return err
}
