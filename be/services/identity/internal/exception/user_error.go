package exception

import "errors"

var (
	ErrUserInvalidInput = errors.New("user: invalid input")
	ErrUserEmailTaken   = errors.New("user: email already registered")
	ErrUserNotFound     = errors.New("user: not found")
)
