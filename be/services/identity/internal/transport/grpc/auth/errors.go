package authgrpc

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"google.golang.org/grpc/codes"
)

var errorMapper = grpcError.NewErrorMapper(
	codes.Internal,
	"identity operation failed",
	grpcError.MatchError(auth.ErrInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(auth.ErrEmailTaken, codes.AlreadyExists, "email already registered"),
	grpcError.MatchError(auth.ErrInvalidCredentials, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(auth.ErrInvalidToken, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(auth.ErrUserInactive, codes.PermissionDenied, "user inactive"),
	grpcError.MatchError(auth.ErrUserNotFound, codes.NotFound, "user not found"),
)
