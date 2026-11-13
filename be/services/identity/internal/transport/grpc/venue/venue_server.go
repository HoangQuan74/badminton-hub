package venuegrpc

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/venue"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type VenueServer struct {
	identityV1.UnimplementedIdentityVenueServiceServer
	service *venue.Service
	auth    *auth.Service
}

func NewVenueServer(service *venue.Service, authentication *auth.Service) *VenueServer {
	return &VenueServer{
		service: service,
		auth:    authentication,
	}
}

func (s *VenueServer) actor(ctx context.Context, token string, ownerOnly bool) (string, error) {
	return identityGRPC.RequireTenantActor(
		ctx,
		s.auth,
		token,
		ownerOnly,
	)
}
