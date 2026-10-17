package identitytest

import (
	"testing"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/security"
)

func TestPasswordRoundTrip(t *testing.T) {
	hash, err := security.HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("security.HashPassword() error = %v", err)
	}
	if !security.ComparePassword(hash, "correct-horse-battery-staple") {
		t.Fatal("security.ComparePassword() rejected the correct password")
	}
	if security.ComparePassword(hash, "wrong-password") {
		t.Fatal("security.ComparePassword() accepted a wrong password")
	}
}
