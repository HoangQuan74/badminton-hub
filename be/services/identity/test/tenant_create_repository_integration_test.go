package identitytest

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	tenantPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/tenant"
	userPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/user"
)

func tenantCreateTestUser(t *testing.T, pool *pgxpool.Pool) model.User {
	t.Helper()
	service := userDomain.NewRegistrationService(userPostgres.NewRepository(pool), auth.NewTokenManager("owner-test-secret", time.Minute, time.Hour))
	user, _, err := service.RegisterUser(context.Background(), userDomain.RegisterInput{Email: uuid.NewString() + "@example.test", Password: "test-password", FullName: "Owner test"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = pool.Exec(ctx, `DELETE FROM identity.organization_members WHERE user_id=$1`, user.ID)
		_, _ = pool.Exec(ctx, `DELETE FROM identity.organizations WHERE id IN (SELECT (response->'organization'->>'ID')::uuid FROM identity.onboard_requests WHERE actor_id=$1)`, user.ID)
		_, _ = pool.Exec(ctx, `DELETE FROM identity.users WHERE id=$1`, user.ID)
	})
	return user
}

func tenantCreateInput() tenant.CreateOrganizationInput {
	return tenant.CreateOrganizationInput{Name: "Smash Center", Slug: "tenantCreate-" + uuid.NewString(), IdempotencyKey: uuid.NewString()}
}

func TestTenantCreatePostgresReplayAndAuthorization(t *testing.T) {
	pool := staffTestPool(t)
	user := tenantCreateTestUser(t, pool)
	service := tenant.NewService(tenantPostgres.NewRepository(pool))
	input := tenantCreateInput()
	ctx := context.Background()
	first, err := service.CreateOrganization(ctx, user.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	if first.Replayed || first.Organization.Status != "PENDING_VERIFICATION" || first.Membership.Status != "ACTIVE" || first.User.RoleName != constant.RoleTenantOwner {
		t.Fatalf("unexpected result: %+v", first)
	}
	if remaining := time.Until(first.ExpiresAt); remaining < 23*time.Hour || remaining > 25*time.Hour {
		t.Fatalf("expiry=%v", remaining)
	}
	input.Name = "  Smash Center  "
	input.Slug = strings.ToUpper(input.Slug)
	replay, err := service.CreateOrganization(ctx, user.ID, input)
	if err != nil || !replay.Replayed || !replay.ExpiresAt.Equal(first.ExpiresAt) {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	firstJSON, _ := json.Marshal(first)
	replayJSON, _ := json.Marshal(replay)
	if string(firstJSON) != string(replayJSON) || strings.Contains(string(firstJSON), "PasswordHash") {
		t.Fatal("snapshot changed or contains credentials")
	}
	input.Name = "Another name"
	if _, err := service.CreateOrganization(ctx, user.ID, input); !errors.Is(err, identityException.ErrTenantKeyConflict) {
		t.Fatalf("key mismatch=%v", err)
	}
	input.Name = "Smash Center"
	input.IdempotencyKey = uuid.NewString()
	if _, err := service.CreateOrganization(ctx, user.ID, input); !errors.Is(err, identityException.ErrTenantSlugConflict) {
		t.Fatalf("slug conflict=%v", err)
	}
	_, err = pool.Exec(ctx, `UPDATE identity.users SET status='suspended' WHERE id=$1`, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	input.IdempotencyKey = firstKey(t, pool, user.ID)
	if _, err := service.CreateOrganization(ctx, user.ID, input); !errors.Is(err, identityException.ErrTenantForbidden) {
		t.Fatalf("suspended replay=%v", err)
	}
	if _, err := service.CreateOrganization(ctx, uuid.NewString(), tenantCreateInput()); !errors.Is(err, identityException.ErrTenantUnauthorized) {
		t.Fatalf("missing actor=%v", err)
	}
}

func firstKey(t *testing.T, pool *pgxpool.Pool, actor string) string {
	t.Helper()
	var key string
	if err := pool.QueryRow(context.Background(), `SELECT idempotency_key::text FROM identity.onboard_requests WHERE actor_id=$1 LIMIT 1`, actor).Scan(&key); err != nil {
		t.Fatal(err)
	}
	return key
}

func TestTenantCreatePostgresConcurrentReplayAndSlugRace(t *testing.T) {
	pool := staffTestPool(t)
	service := tenant.NewService(tenantPostgres.NewRepository(pool))
	user := tenantCreateTestUser(t, pool)
	input := tenantCreateInput()
	var wait sync.WaitGroup
	outputs := make(chan tenant.CreateOrganizationOutput, 8)
	errorsOut := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := service.CreateOrganization(context.Background(), user.ID, input)
			outputs <- result
			errorsOut <- err
		}()
	}
	wait.Wait()
	close(outputs)
	close(errorsOut)
	fresh := 0
	var organizationID string
	for err := range errorsOut {
		if err != nil {
			t.Fatal(err)
		}
	}
	for result := range outputs {
		if !result.Replayed {
			fresh++
		}
		if organizationID != "" && organizationID != result.Organization.ID {
			t.Fatal("duplicate organizations")
		}
		organizationID = result.Organization.ID
	}
	if fresh != 1 {
		t.Fatalf("fresh commits=%d", fresh)
	}

	// Two different actors/keys race for one globally unique slug.
	firstUser, secondUser := tenantCreateTestUser(t, pool), tenantCreateTestUser(t, pool)
	left, right := tenantCreateInput(), tenantCreateInput()
	right.Slug = left.Slug
	raceErrors := make(chan error, 2)
	for i, actor := range []string{firstUser.ID, secondUser.ID} {
		current := []tenant.CreateOrganizationInput{left, right}[i]
		go func() { _, err := service.CreateOrganization(context.Background(), actor, current); raceErrors <- err }()
	}
	successes, conflicts := 0, 0
	for i := 0; i < 2; i++ {
		err := <-raceErrors
		if err == nil {
			successes++
		} else if errors.Is(err, identityException.ErrTenantSlugConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	var owners int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM identity.users WHERE id IN ($1,$2) AND role_id IS NOT NULL`, firstUser.ID, secondUser.ID).Scan(&owners); err != nil || owners != 1 {
		t.Fatalf("partial role mutation: owners=%d err=%v", owners, err)
	}
}

func TestTenantCreatePostgresRollsBackWhenReplayPersistenceFails(t *testing.T) {
	pool := staffTestPool(t)
	user := tenantCreateTestUser(t, pool)
	ctx := context.Background()
	function := "reject_tenantCreate_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	ddl := `CREATE FUNCTION identity.` + function + `() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.actor_id='` + user.ID + `'::uuid THEN RAISE EXCEPTION 'forced replay persistence failure'; END IF; RETURN NEW; END $$; CREATE TRIGGER ` + function + ` BEFORE INSERT ON identity.onboard_requests FOR EACH ROW EXECUTE FUNCTION identity.` + function + `();`
	if _, err := pool.Exec(ctx, ddl); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		_, _ = pool.Exec(ctx, `DROP TRIGGER IF EXISTS `+function+` ON identity.onboard_requests; DROP FUNCTION IF EXISTS identity.`+function+`();`)
	}
	t.Cleanup(cleanup)
	service := tenant.NewService(tenantPostgres.NewRepository(pool))
	input := tenantCreateInput()
	if _, err := service.CreateOrganization(ctx, user.ID, input); err == nil || errors.Is(err, identityException.ErrTenantSlugConflict) {
		t.Fatalf("unexpected failure mapping: %v", err)
	}
	var roleCount, organizations, memberships, keys int
	checks := []struct {
		query  string
		arg    any
		target *int
	}{
		{`SELECT count(*) FROM identity.users WHERE id=$1 AND role_id IS NOT NULL`, user.ID, &roleCount},
		{`SELECT count(*) FROM identity.organizations WHERE slug=$1`, input.Slug, &organizations},
		{`SELECT count(*) FROM identity.organization_members WHERE user_id=$1`, user.ID, &memberships},
		{`SELECT count(*) FROM identity.onboard_requests WHERE actor_id=$1`, user.ID, &keys},
	}
	for _, check := range checks {
		if err := pool.QueryRow(ctx, check.query, check.arg).Scan(check.target); err != nil {
			t.Fatal(err)
		}
	}
	if roleCount+organizations+memberships+keys != 0 {
		t.Fatalf("partial writes: %d %d %d %d", roleCount, organizations, memberships, keys)
	}
	cleanup()
	if _, err := service.CreateOrganization(ctx, user.ID, input); err != nil {
		t.Fatalf("failed request consumed key: %v", err)
	}
}

func TestTenantCreatePostgresDeadlineAndExpiredKey(t *testing.T) {
	pool := staffTestPool(t)
	user := tenantCreateTestUser(t, pool)
	service := tenant.NewService(tenantPostgres.NewRepository(pool))
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT id FROM identity.users WHERE id=$1 FOR UPDATE`, user.ID); err != nil {
		t.Fatal(err)
	}
	input := tenantCreateInput()
	limited, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel()
	if _, err := service.CreateOrganization(limited, user.ID, input); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lock timeout=%v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	first, err := service.CreateOrganization(ctx, user.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE identity.onboard_requests SET expires_at=clock_timestamp()-INTERVAL '1 second' WHERE actor_id=$1`, user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateOrganization(ctx, user.ID, input); !errors.Is(err, identityException.ErrTenantSlugConflict) {
		t.Fatalf("expired replay=%v", err)
	}
	if first.Organization.ID == "" {
		t.Fatal("missing organization")
	}
}

func TestTenantCreatePostgresKeyScopeAndRoleRestrictions(t *testing.T) {
	pool := staffTestPool(t)
	service := tenant.NewService(tenantPostgres.NewRepository(pool))
	ctx := context.Background()
	firstUser, secondUser := tenantCreateTestUser(t, pool), tenantCreateTestUser(t, pool)
	left, right := tenantCreateInput(), tenantCreateInput()
	right.IdempotencyKey = left.IdempotencyKey
	if _, err := service.CreateOrganization(ctx, firstUser.ID, left); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateOrganization(ctx, secondUser.ID, right); err != nil {
		t.Fatalf("keys leaked across actors: %v", err)
	}
	for _, role := range []string{constant.RoleTenantStaff, constant.RoleVenueStaff} {
		staff := tenantCreateTestUser(t, pool)
		if _, err := pool.Exec(ctx, `UPDATE identity.users SET role_id=(SELECT id FROM identity.roles WHERE name=$2) WHERE id=$1`, staff.ID, role); err != nil {
			t.Fatal(err)
		}
		if _, err := service.CreateOrganization(ctx, staff.ID, tenantCreateInput()); !errors.Is(err, identityException.ErrTenantRoleConflict) {
			t.Fatalf("role %s: %v", role, err)
		}
	}
	admin := tenantCreateTestUser(t, pool)
	if _, err := pool.Exec(ctx, `UPDATE identity.users SET is_admin=TRUE WHERE id=$1`, admin.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateOrganization(ctx, admin.ID, tenantCreateInput()); !errors.Is(err, identityException.ErrTenantForbidden) {
		t.Fatalf("admin: %v", err)
	}
}
