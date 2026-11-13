package cataloggrpc

import (
	"context"
	"strings"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/constant"
	catalogException "github.com/kainguyen/badminton-hub/be/services/catalog/internal/exception"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// tenantActor verifies the bearer token with Identity; request actor IDs are not trusted.
func (s *CatalogServer) tenantActor(ctx context.Context) (*identityV1.UserResponse, error) {
	incoming, _ := metadata.FromIncomingContext(ctx)
	values := incoming.Get("authorization")
	if len(values) != 1 {
		return nil, catalogException.ErrGRPCAuthenticationRequired
	}
	scheme, token, found := strings.Cut(values[0], " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return nil, catalogException.ErrGRPCAuthenticationRequired
	}
	if s.identity == nil {
		return nil, catalogException.ErrGRPCIdentityUnavailable
	}
	identityCtx, cancel := context.WithTimeout(ctx, constant.IdentityRPCTimeout)
	defer cancel()
	user, err := s.identity.GetCurrentUser(
		identityCtx,
		&identityV1.GetCurrentUserRequest{
			AccessToken: strings.TrimSpace(token),
		},
		grpc.WaitForReady(false),
	)
	if err != nil {
		return nil, err
	}
	if user.GetId() == "" {
		return nil, catalogException.ErrGRPCInvalidCredentials
	}
	return user, nil
}

func (s *CatalogServer) tenantActorID(ctx context.Context) (string, error) {
	user, err := s.tenantActor(ctx)
	if err != nil {
		return "", err
	}
	return user.GetId(), nil
}

func (s *CatalogServer) tenantOwnerID(ctx context.Context) (string, error) {
	user, err := s.tenantActor(ctx)
	if err != nil {
		return "", err
	}
	for _, role := range user.GetRoles() {
		if role == constant.RoleTenantOwner {
			return user.GetId(), nil
		}
	}
	return "", catalogException.ErrGRPCTenantOwnerRequired
}
