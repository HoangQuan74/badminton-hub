package identitytest

import (
	"errors"
	identityAuth "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"testing"
	"time"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	manager := identityAuth.NewTokenManager("01234567890123456789012345678901", 15*time.Minute, time.Hour)
	token, _, err := manager.NewAccessToken("user-123")
	if err != nil {
		t.Fatalf("NewAccessToken() error = %v", err)
	}
	userID, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}
	if userID != "user-123" {
		t.Fatalf("user ID = %q, want user-123", userID)
	}
}

func TestParseAccessTokenRejectsInvalidToken(t *testing.T) {
	manager := identityAuth.NewTokenManager("01234567890123456789012345678901", 15*time.Minute, time.Hour)
	_, err := manager.ParseAccessToken("not-a-token")
	if !errors.Is(err, identityAuth.ErrInvalidToken) {
		t.Fatalf("error = %v, want identityAuth.ErrInvalidToken", err)
	}
}

func TestRefreshTokenHash(t *testing.T) {
	manager := identityAuth.NewTokenManager("01234567890123456789012345678901", 15*time.Minute, time.Hour)
	first, firstHash, _, err := manager.NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	second, secondHash, _, err := manager.NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("refresh tokens must be random")
	}
	if string(firstHash) != string(identityAuth.HashRefreshToken(first)) {
		t.Fatal("first hash mismatch")
	}
	if string(firstHash) == string(secondHash) {
		t.Fatal("different tokens must not share a hash")
	}
}
