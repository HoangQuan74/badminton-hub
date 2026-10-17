package realtimegrpc

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// ValidateAccessToken verifies an Identity access token signed with the shared secret.
func ValidateAccessToken(raw string, secret []byte) bool {
	claims := new(jwt.RegisteredClaims)
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}
		return secret, nil
	})
	return err == nil && token.Valid && claims.Subject != ""
}
