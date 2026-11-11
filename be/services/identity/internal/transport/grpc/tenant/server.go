package tenantgrpc

import (
	"github.com/jackc/pgx/v5/pgxpool"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	authpostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/auth"
	tenantpostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/tenant"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type Server struct {
	identityV1.UnimplementedIdentityTenantServiceServer
	service *tenant.Service
	auth    *auth.Service
}

func NewServer(cfg identityConfig.Config, pool *pgxpool.Pool) *Server {
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	return &Server{
		service: tenant.NewService(tenantpostgres.New(pool)),
		auth:    auth.NewService(authpostgres.NewAuthRepository(pool), tokens),
	}
}
