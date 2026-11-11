package authgrpc

import (
	"github.com/jackc/pgx/v5/pgxpool"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	authPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/auth"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type AdminServer struct {
	identityV1.UnimplementedIdentityAdminAuthServiceServer
	service    *auth.AdminService
	authorizer *identityGRPC.Authorizer
}

func NewAdminServer(cfg identityConfig.Config, pool *pgxpool.Pool) *AdminServer {
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	return &AdminServer{
		service:    auth.NewAdminService(authPostgres.NewAuthRepository(pool), tokens),
		authorizer: identityGRPC.NewAuthorizer(cfg, pool),
	}
}
