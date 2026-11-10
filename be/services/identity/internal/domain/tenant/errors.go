package tenant

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid tenant input")
	ErrForbidden         = errors.New("tenant access forbidden")
	ErrNotFound          = errors.New("tenant resource not found")
	ErrVenueSlugConflict = errors.New("venue slug already exists")
	ErrSlugConflict      = errors.New("organization slug already exists")
	ErrStateConflict     = errors.New("organization state does not allow this operation")
)
