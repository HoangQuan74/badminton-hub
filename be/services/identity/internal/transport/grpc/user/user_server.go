package usergrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type UserServer struct {
	identityV1.UnimplementedIdentityUserServiceServer
	service    *user.Service
	authorizer *identityGRPC.Authorizer
}

func NewUserServer(service *user.Service, authorizer *identityGRPC.Authorizer) *UserServer {
	return &UserServer{
		service:    service,
		authorizer: authorizer,
	}
}
