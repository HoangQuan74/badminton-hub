package realtimegrpc

import (
	"github.com/golang-jwt/jwt/v5"
	socketException "github.com/kainguyen/badminton-hub/be/services/socket/internal/exception"
)

// ValidateAccessToken verifies an Identity access token signed with the shared secret.
func ValidateAccessToken(raw string, secret []byte) bool {
	claims := new(jwt.RegisteredClaims)
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, socketException.ErrInvalidSigningMethod
		}
		return secret, nil
	})
	return err == nil && token.Valid && claims.Subject != ""
}
