package tenantgrpc

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// actor derives identity from the verified token and the current user record.
func (s *Server) actor(ctx context.Context, token string, ownerOnly bool) (string, error) {
	id, err := s.auth.ParseAccessToken(token)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid credentials")
	}
	user, err := s.auth.Me(ctx, id)
	if err != nil {
		if err == auth.ErrUserNotFound {
			return "", status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return "", status.Error(codes.Unavailable, "identity unavailable")
	}
	if user.Status != model.StatusActive || !auth.IsAllowedPortal(user, "tenant") {
		return "", status.Error(codes.PermissionDenied, "tenant access denied")
	}
	if ownerOnly && user.RoleName != "tenant_owner" {
		return "", status.Error(codes.PermissionDenied, "tenant owner required")
	}
	return user.ID, nil
}
