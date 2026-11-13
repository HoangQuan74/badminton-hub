package exception

import "errors"

var (
	ErrJWTSecretTooShort = errors.New("SOCKET_JWT_SECRET must contain at least 32 characters")
)
