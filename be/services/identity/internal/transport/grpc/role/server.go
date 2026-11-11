package rolegrpc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/role"
	rolePostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/role"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type Handler struct {
	identityV1.UnimplementedIdentityRoleServiceServer
	service    *role.Service
	authorizer *identityGRPC.Authorizer
}

func NewHandler(cfg identityConfig.Config, pool *pgxpool.Pool) *Handler {
	return &Handler{
		service:    role.NewService(rolePostgres.NewRepository(pool)),
		authorizer: identityGRPC.NewAuthorizer(cfg, pool),
	}
}

func (s *Handler) authorizeToken(ctx context.Context, accessToken string) error {
	_, err := s.authorizer.RequireAdmin(ctx, accessToken)
	return err
}
