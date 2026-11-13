package exception

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"google.golang.org/grpc/codes"
)

var TenantGRPCMapper = grpcError.NewErrorMapper(
	codes.Internal, "identity tenant operation failed",
	grpcError.MatchError(ErrAuthInvalidToken, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(ErrStaffNotFound, codes.Unauthenticated, "invalid credentials"),
	grpcError.MatchError(ErrStaffForbidden, codes.PermissionDenied, "permission denied"),
	grpcError.MatchError(ErrTenantRoleConflict, codes.FailedPrecondition, "staff role conflict"),
	grpcError.MatchError(ErrTenantInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(ErrTenantForbidden, codes.PermissionDenied, "tenant access forbidden"),
	grpcError.MatchError(ErrTenantNotFound, codes.NotFound, "tenant resource not found"),
	grpcError.MatchError(ErrTenantSlugConflict, codes.AlreadyExists, "organization slug already exists"),
	grpcError.MatchError(ErrTenantVenueSlugConflict, codes.AlreadyExists, "venue slug already exists"),
	grpcError.MatchError(ErrTenantStateConflict, codes.FailedPrecondition, "organization state conflict"),
)
