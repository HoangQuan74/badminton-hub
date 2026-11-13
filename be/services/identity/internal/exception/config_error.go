package exception

import "errors"

var (
	ErrJWTSecretTooShort = errors.New("IDENTITY_JWT_SECRET must contain at least 32 characters")
)
