package catalog

import "errors"

var (
	ErrOrganizationSlugConflict  = errors.New("organization slug already exists")
	ErrOrganizationStateConflict = errors.New("organization state does not allow this operation")
	ErrUnauthenticated           = errors.New("authentication required")
	ErrIdentityUnavailable       = errors.New("identity unavailable")
	ErrImageStorage              = errors.New("image storage unavailable")
	ErrInvalidInput              = errors.New("invalid catalog input")
	ErrForbidden                 = errors.New("catalog access forbidden")
	ErrNotFound                  = errors.New("catalog resource not found")
	ErrVenueSlugConflict         = errors.New("venue slug already exists")
	ErrCourtNameConflict         = errors.New("court name already exists")
	ErrUnsupportedStatus         = errors.New("unsupported court status")
)
