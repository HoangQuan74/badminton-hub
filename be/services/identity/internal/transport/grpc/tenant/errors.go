package tenantgrpc

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	"google.golang.org/grpc/codes"
)

var errorMapper = grpcError.NewErrorMapper(
	codes.Internal, "identity tenant operation failed",
	grpcError.MatchError(tenant.ErrInvalidInput, codes.InvalidArgument, "invalid request"),
	grpcError.MatchError(tenant.ErrForbidden, codes.PermissionDenied, "tenant access forbidden"),
	grpcError.MatchError(tenant.ErrNotFound, codes.NotFound, "tenant resource not found"),
	grpcError.MatchError(tenant.ErrSlugConflict, codes.AlreadyExists, "organization slug already exists"),
	grpcError.MatchError(tenant.ErrVenueSlugConflict, codes.AlreadyExists, "venue slug already exists"),
	grpcError.MatchError(tenant.ErrStateConflict, codes.FailedPrecondition, "organization state conflict"),
)
