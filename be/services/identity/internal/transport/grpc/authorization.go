package grpc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	staffDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	staffPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/staff"
)

type Authorizer struct {
	tokens *auth.TokenManager
	staff  *staffDomain.Service
}

func NewAuthorizer(cfg identityConfig.Config, pool *pgxpool.Pool) *Authorizer {
	return &Authorizer{
		tokens: auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL),
		staff:  staffDomain.NewService(staffPostgres.NewRepository(pool)),
	}
}

func (a *Authorizer) RequireAdmin(ctx context.Context, accessToken string) (model.Staff, error) {
	staffID, err := a.tokens.ParseAccessToken(accessToken)
	if err != nil {
		return model.Staff{}, auth.ErrInvalidToken
	}
	value, err := a.staff.Get(ctx, staffID)
	if err != nil {
		return model.Staff{}, err
	}
	if !value.IsAdmin {
		return model.Staff{}, staffDomain.ErrForbidden
	}
	return value, nil
}
