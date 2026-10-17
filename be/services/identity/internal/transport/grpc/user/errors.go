package usergrpc

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	staffDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	"google.golang.org/grpc/codes"
)

var errorMapper = grpcError.NewErrorMapper(
	codes.Internal,
	"identity user operation failed",
	grpcError.MatchError(auth.ErrInvalidToken, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(staffDomain.ErrForbidden, codes.PermissionDenied, "permission denied"),
	grpcError.MatchError(userDomain.ErrInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(userDomain.ErrEmailTaken, codes.AlreadyExists, "email already registered"),
	grpcError.MatchError(userDomain.ErrNotFound, codes.NotFound, "user not found"),
)
