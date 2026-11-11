package authgrpc

import (
	"github.com/jackc/pgx/v5/pgxpool"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	authPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/auth"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type TenantServer struct {
	identityV1.UnimplementedIdentityTenantAuthServiceServer
	service *auth.TenantService
}

func NewTenantServer(cfg identityConfig.Config, pool *pgxpool.Pool) *TenantServer {
	// Build tenant authentication with a fixed portal boundary.
	repository := authPostgres.NewAuthRepository(pool)
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	return &TenantServer{
		service: auth.NewTenantService(repository, tokens),
	}
}
