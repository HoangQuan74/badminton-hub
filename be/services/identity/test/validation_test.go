package identitytest

import (
	"testing"

	authDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

func TestValidationPackage(t *testing.T) {
	valid := authDomain.RegisterInput{
		Email:    "user@example.com",
		Password: "password123",
		FullName: "User",
	}
	if err := validation.Struct(valid); err != nil {
		t.Fatalf("validation.Struct() unexpected error: %v", err)
	}

	invalid := authDomain.RegisterInput{
		Email:    "invalid",
		Password: "short",
	}
	if err := validation.Struct(invalid); err == nil {
		t.Fatal("validation.Struct() error = nil, want validation error")
	}
}
