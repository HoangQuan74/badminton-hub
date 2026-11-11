package authgrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func tokenPair(pair auth.TokenPair) *identityV1.TokenPair {
	// Map domain tokens to the gRPC token response.
	return &identityV1.TokenPair{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        "Bearer",
		AccessExpiresAt:  pair.AccessExpiresAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		RefreshExpiresAt: pair.RefreshExpiresAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
	}
}
