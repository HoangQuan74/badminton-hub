package staffgrpc

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type StaffServer struct {
	identityV1.UnimplementedIdentityStaffServiceServer
	service     *staff.Service
	memberships *staff.MembershipService
	tenantAuth  *auth.Service
	authorizer  *identityGRPC.Authorizer
}

func NewStaffServer(service *staff.Service, memberships *staff.MembershipService, authentication *auth.Service, authorizer *identityGRPC.Authorizer) *StaffServer {
	return &StaffServer{
		service:     service,
		memberships: memberships,
		tenantAuth:  authentication,
		authorizer:  authorizer,
	}
}

func (s *StaffServer) tenantActor(ctx context.Context, token string, ownerOnly bool) (string, error) {
	return identityGRPC.RequireTenantActor(
		ctx,
		s.tenantAuth,
		token,
		ownerOnly,
	)
}
