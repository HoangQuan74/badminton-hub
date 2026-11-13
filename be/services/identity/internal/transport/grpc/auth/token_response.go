package authgrpc

import (
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func tokenPair(pair model.TokenPair) *identityV1.TokenPair {
	return mapTokenPair(pair, time.RFC3339Nano)
}

func mapTokenPair(pair model.TokenPair, timestampLayout string) *identityV1.TokenPair {
	// Map token fields once while preserving each endpoint's timestamp precision.
	return &identityV1.TokenPair{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        "Bearer",
		AccessExpiresAt:  pair.AccessExpiresAt.UTC().Format(timestampLayout),
		RefreshExpiresAt: pair.RefreshExpiresAt.UTC().Format(timestampLayout),
	}
}
