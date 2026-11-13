package identitytest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/security"
)

type tenantAuthRepository struct {
	auth.AuthRepository
	user     model.User
	sessions int
}

func (r *tenantAuthRepository) FindCredentialsByEmail(context.Context, string) (model.User, error) {
	return r.user, nil
}

func (r *tenantAuthRepository) GetPrincipal(context.Context, string) (model.User, error) {
	return r.user, nil
}

func (r *tenantAuthRepository) CreateSession(context.Context, model.Session) error {
	r.sessions++
	return nil
}

func TestTenantAuthAccountBoundary(t *testing.T) {
	hash, err := security.HashPassword("test-password")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		role    string
		admin   bool
		allowed bool
	}{
		{name: "owner", role: constant.RoleTenantOwner, allowed: true},
		{name: "tenant staff", role: constant.RoleTenantStaff, allowed: true},
		{name: "venue staff", role: constant.RoleVenueStaff, allowed: true},
		{name: constant.RoleCustomer},
		{name: "admin", admin: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &tenantAuthRepository{user: model.User{
				ID: "test-user", Email: "test@example.com", PasswordHash: hash,
				Status: constant.AccountStatusActive, RoleName: test.role, IsAdmin: test.admin,
			}}
			tokens := auth.NewTokenManager("test-secret", time.Minute, time.Hour)
			service := auth.NewService(repository, tokens, tenant.CanAccessPortal)
			// A caller cannot select the user portal to bypass the tenant boundary.
			_, _, err := service.Login(context.Background(), auth.LoginInput{
				Email: "test@example.com", Password: "test-password",
			})
			if test.allowed {
				if err != nil || repository.sessions != 1 {
					t.Fatalf("login: %v, sessions: %d", err, repository.sessions)
				}
			} else if !errors.Is(err, identityException.ErrAuthWrongPortal) || repository.sessions != 0 {
				t.Fatalf("unexpected access: %v, sessions: %d", err, repository.sessions)
			}
			access, _, err := tokens.NewAccessToken(repository.user.ID)
			if err != nil {
				t.Fatal(err)
			}
			_, err = service.CurrentPrincipal(context.Background(), access)
			if test.allowed && err != nil {
				t.Fatalf("me: %v", err)
			}
			if !test.allowed && !errors.Is(err, identityException.ErrAuthWrongPortal) {
				t.Fatalf("me unexpectedly allowed: %v", err)
			}
		})
	}
}
