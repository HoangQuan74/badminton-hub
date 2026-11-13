package exception

import "errors"

var (
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrInvalidAccessToken   = errors.New("invalid access token")
)
