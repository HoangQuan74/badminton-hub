package authgrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type AdminServer struct {
	identityV1.UnimplementedIdentityAdminAuthServiceServer
	service    *staff.AccountService
	authorizer *identityGRPC.Authorizer
}

func NewAdminServer(service *staff.AccountService, authorizer *identityGRPC.Authorizer) *AdminServer {
	return &AdminServer{
		service:    service,
		authorizer: authorizer,
	}
}
