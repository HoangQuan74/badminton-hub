package cataloggrpc

import (
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	"google.golang.org/grpc/codes"
)

var errorMapper = grpcError.NewErrorMapper(
	codes.Internal,
	"catalog operation failed",
	grpcError.MatchError(catalog.ErrUnauthenticated, codes.Unauthenticated, catalog.ErrUnauthenticated.Error()),
	grpcError.MatchError(catalog.ErrIdentityUnavailable, codes.Unavailable, catalog.ErrIdentityUnavailable.Error()),
	grpcError.MatchError(catalog.ErrOrganizationSlugConflict, codes.AlreadyExists, catalog.ErrOrganizationSlugConflict.Error()),
	grpcError.MatchError(catalog.ErrOrganizationStateConflict, codes.FailedPrecondition, catalog.ErrOrganizationStateConflict.Error()),
	grpcError.MatchError(catalog.ErrImageStorage, codes.Unavailable, catalog.ErrImageStorage.Error()),
	grpcError.MatchError(catalog.ErrInvalidInput, codes.InvalidArgument, catalog.ErrInvalidInput.Error()),
	grpcError.MatchError(catalog.ErrUnsupportedStatus, codes.InvalidArgument, catalog.ErrUnsupportedStatus.Error()),
	grpcError.MatchError(catalog.ErrForbidden, codes.PermissionDenied, catalog.ErrForbidden.Error()),
	grpcError.MatchError(catalog.ErrNotFound, codes.NotFound, catalog.ErrNotFound.Error()),
	grpcError.MatchError(catalog.ErrVenueSlugConflict, codes.AlreadyExists, catalog.ErrVenueSlugConflict.Error()),
	grpcError.MatchError(catalog.ErrCourtNameConflict, codes.AlreadyExists, catalog.ErrCourtNameConflict.Error()),
)
