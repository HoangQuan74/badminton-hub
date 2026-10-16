package helper

import "testing"

func TestPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if !ComparePassword(hash, "correct-horse-battery-staple") {
		t.Fatal("ComparePassword() rejected the correct password")
	}
	if ComparePassword(hash, "wrong-password") {
		t.Fatal("ComparePassword() accepted a wrong password")
	}
}
