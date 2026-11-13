package authgrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type UserServer struct {
	identityV1.UnimplementedIdentityAuthServiceServer
	session      *sessionHandler
	registration *userDomain.RegistrationService
}

func NewUserServer(service *auth.Service, registration *userDomain.RegistrationService) *UserServer {
	return &UserServer{
		session:      &sessionHandler{service: service},
		registration: registration,
	}
}
