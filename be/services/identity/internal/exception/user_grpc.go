package exception

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"google.golang.org/grpc/codes"
)

var UserGRPCMapper = grpcError.NewErrorMapper(
	codes.Internal,
	"identity user operation failed",
	grpcError.MatchError(ErrAuthInvalidToken, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(ErrStaffForbidden, codes.PermissionDenied, "permission denied"),
	grpcError.MatchError(ErrUserInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(ErrUserEmailTaken, codes.AlreadyExists, "email already registered"),
	grpcError.MatchError(ErrUserNotFound, codes.NotFound, "user not found"),
)
