package staffgrpc

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	staffDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	"google.golang.org/grpc/codes"
)

var errorMapper = grpcError.NewErrorMapper(
	codes.Internal,
	"identity staff operation failed",
	grpcError.MatchError(auth.ErrInvalidToken, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(staffDomain.ErrForbidden, codes.PermissionDenied, "permission denied"),
	grpcError.MatchError(staffDomain.ErrInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(staffDomain.ErrEmailTaken, codes.AlreadyExists, "email already registered"),
	grpcError.MatchError(staffDomain.ErrNotFound, codes.NotFound, "staff not found"),
)
