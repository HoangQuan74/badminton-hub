package tenantgrpc

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type TenantServer struct {
	identityV1.UnimplementedIdentityTenantServiceServer
	service    *tenant.Service
	auth       *auth.Service
	tokens     *auth.TokenManager
	authorizer *identityGRPC.Authorizer
}

func NewTenantServer(service *tenant.Service, authentication *auth.Service, tokens *auth.TokenManager, authorizer *identityGRPC.Authorizer) *TenantServer {
	return &TenantServer{
		service:    service,
		auth:       authentication,
		tokens:     tokens,
		authorizer: authorizer,
	}
}

func (s *TenantServer) actor(ctx context.Context, token string, ownerOnly bool) (string, error) {
	return identityGRPC.RequireTenantActor(
		ctx,
		s.auth,
		token,
		ownerOnly,
	)
}
