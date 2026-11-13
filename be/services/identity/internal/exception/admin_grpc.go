package exception

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"google.golang.org/grpc/codes"
)

var AdminGRPCMapper = grpcError.NewErrorMapper(
	codes.Internal, "identity staff operation failed",
	grpcError.MatchError(ErrAuthInvalidToken, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(ErrAuthInvalidCredentials, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(ErrAuthUserInactive, codes.PermissionDenied, "user inactive"),
	grpcError.MatchError(ErrStaffForbidden, codes.PermissionDenied, "permission denied"),
	grpcError.MatchError(ErrStaffInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(ErrStaffNotFound, codes.NotFound, "staff not found"),
)
