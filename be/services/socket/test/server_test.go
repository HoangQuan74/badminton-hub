package test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	realtimeGRPC "github.com/kainguyen/badminton-hub/be/services/socket/internal/transport/grpc/realtime"
)

func TestValidAccessToken(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "user-id",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	if !realtimeGRPC.ValidateAccessToken(token, secret) {
		t.Fatal("valid token rejected")
	}
	if realtimeGRPC.ValidateAccessToken(token, []byte("wrong-secret")) {
		t.Fatal("token accepted with wrong secret")
	}
}
