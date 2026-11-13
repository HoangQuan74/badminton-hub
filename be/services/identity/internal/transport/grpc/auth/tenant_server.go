package authgrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type TenantServer struct {
	identityV1.UnimplementedIdentityTenantAuthServiceServer
	session *sessionHandler
}

func NewTenantServer(service *auth.Service) *TenantServer {
	return &TenantServer{session: &sessionHandler{service: service}}
}
