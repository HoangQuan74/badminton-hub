package exception

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"google.golang.org/grpc/codes"
)

var RoleGRPCMapper = grpcError.NewErrorMapper(
	codes.Internal, "identity staff operation failed",
	grpcError.MatchError(ErrAuthInvalidToken, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(ErrStaffForbidden, codes.PermissionDenied, "permission denied"),
	grpcError.MatchError(ErrStaffInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(ErrStaffNotFound, codes.NotFound, "staff not found"),
	grpcError.MatchError(ErrRoleInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(ErrRoleNotFound, codes.NotFound, "staff not found"),
)
