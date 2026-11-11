package rolegrpc

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/role"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	"google.golang.org/grpc/codes"
)

var errorMapper = grpcError.NewErrorMapper(
	codes.Internal, "identity staff operation failed",
	grpcError.MatchError(auth.ErrInvalidToken, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(staff.ErrForbidden, codes.PermissionDenied, "permission denied"),
	grpcError.MatchError(staff.ErrInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(staff.ErrNotFound, codes.NotFound, "staff not found"),
	grpcError.MatchError(role.ErrInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(role.ErrNotFound, codes.NotFound, "staff not found"),
)
