package exception

import "errors"

var (
	ErrStaffInvalidInput = errors.New("staff: invalid input")
	ErrStaffEmailTaken   = errors.New("staff: email already registered")
	ErrStaffNotFound     = errors.New("staff: not found")
	ErrStaffForbidden    = errors.New("staff: forbidden")
)
