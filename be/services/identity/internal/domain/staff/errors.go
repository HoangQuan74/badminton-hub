package staff

import "errors"

var (
	ErrInvalidInput = errors.New("staff: invalid input")
	ErrEmailTaken   = errors.New("staff: email already registered")
	ErrNotFound     = errors.New("staff: not found")
	ErrForbidden    = errors.New("staff: forbidden")
)
