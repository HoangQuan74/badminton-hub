package exception

import "errors"

var (
	ErrRoleInvalidInput = errors.New("role.invalid_input")
	ErrRoleNotFound     = errors.New("role.not_found")
)
