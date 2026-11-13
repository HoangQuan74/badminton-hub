package identitytest

import (
	"testing"

	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

func TestValidationPackage(t *testing.T) {
	valid := userDomain.RegisterInput{
		Email:    "user@example.com",
		Password: "password123",
		FullName: "User",
	}
	if err := validation.Struct(valid); err != nil {
		t.Fatalf("validation.Struct() unexpected error: %v", err)
	}

	invalid := userDomain.RegisterInput{
		Email:    "invalid",
		Password: "short",
	}
	if err := validation.Struct(invalid); err == nil {
		t.Fatal("validation.Struct() error = nil, want validation error")
	}
}
