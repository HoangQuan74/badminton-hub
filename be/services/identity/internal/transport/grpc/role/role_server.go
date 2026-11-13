package rolegrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/role"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type RoleServer struct {
	identityV1.UnimplementedIdentityRoleServiceServer
	service    *role.Service
	authorizer *identityGRPC.Authorizer
}

func NewRoleServer(service *role.Service, authorizer *identityGRPC.Authorizer) *RoleServer {
	return &RoleServer{
		service:    service,
		authorizer: authorizer,
	}
}
