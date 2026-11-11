package authgrpc

import (
	"github.com/jackc/pgx/v5/pgxpool"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	authPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/auth"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type UserServer struct {
	identityV1.UnimplementedIdentityAuthServiceServer
	service *auth.Service
}

func NewUserServer(cfg identityConfig.Config, pool *pgxpool.Pool) *UserServer {
	// Build the auth repository and token manager dependencies.
	repository := authPostgres.NewAuthRepository(pool)
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	return &UserServer{
		service: auth.NewService(repository, tokens),
	}
}
