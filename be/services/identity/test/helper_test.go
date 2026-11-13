package identitytest

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	postgresHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/helper"
	grpcHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/helper"
)

func TestPostgresOptionalValuesPreserveExplicitEmptyAndZero(t *testing.T) {
	empty := ""
	if postgresHelper.MapOptionalText(nil).Valid {
		t.Fatal("omitted text must remain SQL NULL")
	}
	if value := postgresHelper.MapOptionalText(&empty); !value.Valid || value.String != "" {
		t.Fatal("explicit empty text must remain a valid update value")
	}

	zero := time.Time{}
	if postgresHelper.MapOptionalTimestamp(nil).Valid {
		t.Fatal("omitted date filter must remain SQL NULL")
	}
	if value := postgresHelper.MapOptionalTimestamp(&zero); !value.Valid || !value.Time.IsZero() {
		t.Fatal("explicit zero timestamp must not become an omitted filter")
	}
}

func TestTenantScopeParsePreservesErrorCategories(t *testing.T) {
	actorID := uuid.NewString()
	resourceID := uuid.NewString()
	actor, resource, err := postgresHelper.ParseTenantScope(actorID, resourceID)
	if err != nil || actor.String() != actorID || resource.String() != resourceID {
		t.Fatalf("valid scope: %v", err)
	}

	if _, _, err := postgresHelper.ParseTenantScope("invalid", resourceID); !errors.Is(err, identityException.ErrTenantForbidden) {
		t.Fatalf("invalid actor: %v", err)
	}
	if _, _, err := postgresHelper.ParseTenantScope(actorID, "invalid"); !errors.Is(err, identityException.ErrTenantInvalidInput) {
		t.Fatalf("invalid resource: %v", err)
	}
}

func TestUserResponseHelperKeepsAccessFieldsCallerOwned(t *testing.T) {
	user := model.User{
		ID:           uuid.NewString(),
		PasswordHash: "private-password-hash",
		IsAdmin:      true,
		RoleName:     "private-role",
		CreatedAt:    time.Date(2026, time.November, 13, 10, 0, 0, 123, time.FixedZone("test", 7*60*60)),
	}
	response := grpcHelper.MapUserResponse(user)
	if response.GetIsAdmin() || len(response.GetRoles()) != 0 {
		t.Fatal("base profile mapping must not add access fields to CRUD responses")
	}
	if response.GetCreatedAt() != "2026-11-13T03:00:00.000000123Z" {
		t.Fatalf("timestamp precision changed: %s", response.GetCreatedAt())
	}
}
