package user

import "errors"

var (
	ErrInvalidInput = errors.New("user: invalid input")
	ErrEmailTaken   = errors.New("user: email already registered")
	ErrNotFound     = errors.New("user: not found")
)
