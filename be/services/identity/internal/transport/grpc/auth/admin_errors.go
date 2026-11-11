package authgrpc

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	"google.golang.org/grpc/codes"
)

var adminErrorMapper = grpcError.NewErrorMapper(
	codes.Internal, "identity staff operation failed",
	grpcError.MatchError(auth.ErrInvalidToken, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(auth.ErrInvalidCredentials, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(auth.ErrUserInactive, codes.PermissionDenied, "user inactive"),
	grpcError.MatchError(staff.ErrForbidden, codes.PermissionDenied, "permission denied"),
	grpcError.MatchError(staff.ErrInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(staff.ErrNotFound, codes.NotFound, "staff not found"),
)
