package exception

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"google.golang.org/grpc/codes"
)

var AuthGRPCMapper = grpcError.NewErrorMapper(
	codes.Internal,
	"identity operation failed",
	grpcError.MatchError(ErrAuthInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(ErrAuthEmailTaken, codes.AlreadyExists, "email already registered"),
	grpcError.MatchError(ErrAuthInvalidCredentials, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(ErrAuthInvalidToken, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(ErrAuthUserInactive, codes.PermissionDenied, "user inactive"),
	grpcError.MatchError(ErrAuthWrongPortal, codes.PermissionDenied, "account cannot sign in to this portal"),
	grpcError.MatchError(ErrAuthUserNotFound, codes.NotFound, "user not found"),
)
