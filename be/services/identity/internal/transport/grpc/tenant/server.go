package tenantgrpc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	authpostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/auth"
	tenantpostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/tenant"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	identityV1.UnimplementedIdentityTenantServiceServer
	service *tenant.Service
	auth    *auth.Service
}

func NewServer(cfg identityConfig.Config, pool *pgxpool.Pool) *Server {
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	return &Server{service: tenant.NewService(tenantpostgres.New(pool)), auth: auth.NewService(authpostgres.NewAuthRepository(pool), tokens)}
}

// actor derives identity from the verified token and the current user record.
func (s *Server) actor(ctx context.Context, token string, ownerOnly bool) (string, error) {
	id, err := s.auth.ParseAccessToken(token)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid credentials")
	}
	user, err := s.auth.Me(ctx, id)
	if err != nil {
		if err == auth.ErrUserNotFound {
			return "", status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return "", status.Error(codes.Unavailable, "identity unavailable")
	}
	if user.Status != model.StatusActive || !auth.IsAllowedPortal(user, "tenant") {
		return "", status.Error(codes.PermissionDenied, "tenant access denied")
	}
	if ownerOnly && user.RoleName != "tenant_owner" {
		return "", status.Error(codes.PermissionDenied, "tenant owner required")
	}
	return user.ID, nil
}
