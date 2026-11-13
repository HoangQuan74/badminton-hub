package exception

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IdentityError maps failed tenant-directory calls without classifying them as missing data.
func IdentityError(err error, conflict error) error {
	switch status.Code(err) {
	case codes.Unauthenticated:
		return ErrUnauthenticated
	case codes.PermissionDenied:
		return ErrForbidden
	case codes.InvalidArgument:
		return ErrInvalidInput
	case codes.NotFound:
		return ErrNotFound
	case codes.AlreadyExists:
		if conflict != nil {
			return conflict
		}
	case codes.FailedPrecondition:
		return ErrOrganizationStateConflict
	case codes.Unavailable, codes.DeadlineExceeded:
		return fmt.Errorf("%w: %v", ErrIdentityUnavailable, err)
	}
	return fmt.Errorf("identity tenant request: %w", err)
}
