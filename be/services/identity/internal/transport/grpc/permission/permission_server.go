package permissiongrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/permission"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type PermissionServer struct {
	identityV1.UnimplementedIdentityPermissionServiceServer
	service    *permission.Service
	authorizer *identityGRPC.Authorizer
}

func NewPermissionServer(service *permission.Service, authorizer *identityGRPC.Authorizer) *PermissionServer {
	return &PermissionServer{
		service:    service,
		authorizer: authorizer,
	}
}
