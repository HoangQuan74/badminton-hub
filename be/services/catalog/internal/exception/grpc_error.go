package exception

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"google.golang.org/grpc/codes"
)

var GRPCMapper = grpcError.NewErrorMapper(
	codes.Internal,
	"catalog operation failed",
	grpcError.MatchError(ErrUnauthenticated, codes.Unauthenticated, ErrUnauthenticated.Error()),
	grpcError.MatchError(ErrIdentityUnavailable, codes.Unavailable, ErrIdentityUnavailable.Error()),
	grpcError.MatchError(ErrOrganizationSlugConflict, codes.AlreadyExists, ErrOrganizationSlugConflict.Error()),
	grpcError.MatchError(ErrOrganizationStateConflict, codes.FailedPrecondition, ErrOrganizationStateConflict.Error()),
	grpcError.MatchError(ErrImageStorage, codes.Unavailable, ErrImageStorage.Error()),
	grpcError.MatchError(ErrInvalidInput, codes.InvalidArgument, ErrInvalidInput.Error()),
	grpcError.MatchError(ErrUnsupportedStatus, codes.InvalidArgument, ErrUnsupportedStatus.Error()),
	grpcError.MatchError(ErrForbidden, codes.PermissionDenied, ErrForbidden.Error()),
	grpcError.MatchError(ErrNotFound, codes.NotFound, ErrNotFound.Error()),
	grpcError.MatchError(ErrVenueSlugConflict, codes.AlreadyExists, ErrVenueSlugConflict.Error()),
	grpcError.MatchError(ErrCourtNameConflict, codes.AlreadyExists, ErrCourtNameConflict.Error()),
)
