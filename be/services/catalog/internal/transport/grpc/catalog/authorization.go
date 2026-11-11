package cataloggrpc

import (
	"context"
	"strings"
	"time"

	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// tenantActor verifies the bearer token with Identity; request actor IDs are not trusted.
func (s *Server) tenantActor(ctx context.Context) (*identityV1.UserResponse, error) {
	incoming, _ := metadata.FromIncomingContext(ctx)
	values := incoming.Get("authorization")
	if len(values) != 1 {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	scheme, token, found := strings.Cut(values[0], " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	if s.identity == nil {
		return nil, status.Error(codes.Unavailable, "identity unavailable")
	}
	identityCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	user, err := s.identity.GetCurrentUser(
		identityCtx,
		&identityV1.GetCurrentUserRequest{
			AccessToken: strings.TrimSpace(token),
			Portal:      "tenant",
		},
		grpc.WaitForReady(false),
	)
	if err != nil {
		return nil, err
	}
	if user.GetId() == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	return user, nil
}

func (s *Server) tenantActorID(ctx context.Context) (string, error) {
	user, err := s.tenantActor(ctx)
	if err != nil {
		return "", err
	}
	return user.GetId(), nil
}

func (s *Server) tenantOwnerID(ctx context.Context) (string, error) {
	user, err := s.tenantActor(ctx)
	if err != nil {
		return "", err
	}
	for _, role := range user.GetRoles() {
		if role == "tenant_owner" {
			return user.GetId(), nil
		}
	}
	return "", status.Error(codes.PermissionDenied, "tenant owner required")
}
