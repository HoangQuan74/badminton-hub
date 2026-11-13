package identitytest

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type tenantCreateRepository struct {
	tenant.TenantRepository
	params tenant.CreateOrganizationParams
	calls  int
	err    error
}

func (r *tenantCreateRepository) WithinCreationTransaction(ctx context.Context, execute func(tenant.TenantCreationTransaction) error) error {
	r.calls++
	if r.err != nil {
		return r.err
	}
	return execute(&tenantCreationTransaction{repository: r})
}

type tenantCreationTransaction struct {
	tenant.TenantCreationTransaction
	repository *tenantCreateRepository
}

func (t *tenantCreationTransaction) LockCreator(_ context.Context, actorID string) (model.User, error) {
	return model.User{ID: actorID, Status: constant.AccountStatusActive}, nil
}
func (t *tenantCreationTransaction) FindCreateReplay(context.Context, string, string) (tenant.CreationReplay, bool, error) {
	return tenant.CreationReplay{}, false, nil
}
func (t *tenantCreationTransaction) DeleteExpiredCreateKey(context.Context, string, string) error {
	return nil
}
func (t *tenantCreationTransaction) SetOwnerRole(context.Context, string) error { return nil }
func (t *tenantCreationTransaction) CreateOrganization(_ context.Context, input tenant.CreateOrganizationParams) (model.Organization, error) {
	t.repository.params = input
	return model.Organization{ID: uuid.NewString(), Name: input.Name, Slug: input.Slug, Status: constant.OrganizationStatusPendingVerification}, nil
}
func (t *tenantCreationTransaction) SaveRegistration(context.Context, string, tenant.RegistrationContactInput) error {
	return nil
}
func (t *tenantCreationTransaction) SaveCreateReplay(_ context.Context, input tenant.CreateOrganizationParams, _ tenant.CreateOrganizationOutput) (time.Time, error) {
	t.repository.params = input
	return time.Now().Add(24 * time.Hour), nil
}

func TestTenantCreateValidationAndCanonicalization(t *testing.T) {
	input := tenant.CreateOrganizationInput{Name: " Smash Center ", Slug: " SMASH-CENTER ", IdempotencyKey: uuid.NewString()}
	repository := &tenantCreateRepository{}
	service := tenant.NewService(repository)
	actor := uuid.NewString()
	if _, err := service.CreateOrganization(context.Background(), actor, input); err != nil {
		t.Fatal(err)
	}
	if repository.params.Name != "Smash Center" || repository.params.Slug != "smash-center" {
		t.Fatalf("not canonicalized: %+v", repository.params)
	}
	firstHash := append([]byte(nil), repository.params.RequestHash...)
	input.Name, input.Slug = "Smash Center", "smash-center"
	input.IdempotencyKey = strings.ToUpper(input.IdempotencyKey)
	if _, err := service.CreateOrganization(context.Background(), actor, input); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstHash, repository.params.RequestHash) {
		t.Fatal("equivalent payloads must hash identically")
	}

	for _, test := range []struct {
		name  string
		input tenant.CreateOrganizationInput
		field string
	}{
		{"blank name", tenant.CreateOrganizationInput{Name: "   ", Slug: "valid", IdempotencyKey: input.IdempotencyKey}, "name"},
		{"nul name", tenant.CreateOrganizationInput{Name: "invalid\x00name", Slug: "valid", IdempotencyKey: input.IdempotencyKey}, "name"},
		{"long name", tenant.CreateOrganizationInput{Name: strings.Repeat("界", 121), Slug: "valid", IdempotencyKey: input.IdempotencyKey}, "name"},
		{"empty slug", tenant.CreateOrganizationInput{Name: "Valid", IdempotencyKey: input.IdempotencyKey}, "slug"},
		{"double hyphen", tenant.CreateOrganizationInput{Name: "Valid", Slug: "bad--slug", IdempotencyKey: input.IdempotencyKey}, "slug"},
		{"unicode slug", tenant.CreateOrganizationInput{Name: "Valid", Slug: "sân", IdempotencyKey: input.IdempotencyKey}, "slug"},
		{"long slug", tenant.CreateOrganizationInput{Name: "Valid", Slug: strings.Repeat("a", 101), IdempotencyKey: input.IdempotencyKey}, "slug"},
		{"invalid key", tenant.CreateOrganizationInput{Name: "Valid", Slug: "valid", IdempotencyKey: "bad"}, "idempotency_key"},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := repository.calls
			_, err := service.CreateOrganization(context.Background(), actor, test.input)
			var invalid *identityException.TenantValidationError
			if !errors.As(err, &invalid) || !errors.Is(err, identityException.ErrTenantInvalidInput) || repository.calls != before {
				t.Fatalf("validation error = %v", err)
			}
			found := false
			for _, field := range invalid.Fields {
				if field.Field == test.field {
					found = true
				}
			}
			if !found {
				t.Fatalf("missing field %s: %+v", test.field, invalid.Fields)
			}
		})
	}
	input.Name, input.Slug = strings.Repeat("界", 120), strings.Repeat("a", 100)
	if _, err := service.CreateOrganization(context.Background(), actor, input); err != nil {
		t.Fatalf("boundary values rejected: %v", err)
	}
	if _, err := service.CreateOrganization(context.Background(), "invalid", input); !errors.Is(err, identityException.ErrTenantUnauthorized) {
		t.Fatalf("actor: %v", err)
	}
	repository.err = identityException.ErrTenantUnavailable
	if _, err := service.CreateOrganization(context.Background(), actor, input); !errors.Is(err, identityException.ErrTenantUnavailable) {
		t.Fatalf("dependency error: %v", err)
	}
}

func TestTenantCreateEligibility(t *testing.T) {
	for _, test := range []struct {
		name string
		user model.User
		want error
	}{
		{constant.RoleCustomer, model.User{Status: constant.AccountStatusActive}, nil},
		{"named customer", model.User{Status: constant.AccountStatusActive, RoleName: constant.RoleCustomer}, nil},
		{"owner", model.User{Status: constant.AccountStatusActive, RoleName: constant.RoleTenantOwner}, nil},
		{"suspended", model.User{Status: constant.AccountStatusSuspended}, identityException.ErrTenantForbidden},
		{"admin", model.User{Status: constant.AccountStatusActive, IsAdmin: true}, identityException.ErrTenantForbidden},
		{"staff", model.User{Status: constant.AccountStatusActive, RoleName: constant.RoleTenantStaff}, identityException.ErrTenantRoleConflict},
		{"venue staff", model.User{Status: constant.AccountStatusActive, RoleName: constant.RoleVenueStaff}, identityException.ErrTenantRoleConflict},
		{"custom role", model.User{Status: constant.AccountStatusActive, RoleName: "custom"}, identityException.ErrTenantRoleConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := tenant.CheckActor(test.user); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}
