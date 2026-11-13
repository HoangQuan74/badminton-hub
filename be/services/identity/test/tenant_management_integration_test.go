package identitytest

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/app"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	staffPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/staff"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestTenantRegistrationApprovalVenueAndStaffLifecycle(t *testing.T) {
	// Register real module services against a disposable, fully migrated database.
	pool := staffTestPool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg := identityConfig.Config{JWTSecret: "tenant-management-test-secret-long-enough", AccessTokenTTL: time.Hour, RefreshTokenTTL: time.Hour}
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	server := grpc.NewServer()
	app.RegisterGRPCServices(cfg, pool)(server)
	listener := bufconn.Listen(1024 * 1024)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	connection, err := grpc.NewClient("passthrough:///tenant-management", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	tenantClient := identityV1.NewIdentityTenantServiceClient(connection)
	venueClient := identityV1.NewIdentityVenueServiceClient(connection)
	staffClient := identityV1.NewIdentityStaffServiceClient(connection)
	owner := tenantCreateTestUser(t, pool)
	ownerToken, _, err := tokens.NewAccessToken(owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	adminRepository := staffPostgres.NewRepository(pool)
	admin, err := adminRepository.CreateStaff(ctx, staff.CreateStaffParams{Email: uuid.NewString() + "@example.test", PasswordHash: "unused", FullName: "Reviewer", IsAdmin: true})
	if err != nil {
		t.Fatal(err)
	}
	adminToken, _, err := tokens.NewAccessToken(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	organizationIDs := []string{}
	defer func() {
		cleanup := context.Background()
		for _, id := range organizationIDs {
			_, _ = pool.Exec(cleanup, "DELETE FROM identity.venue_staff WHERE organization_id = $1", id)
			_, _ = pool.Exec(cleanup, "DELETE FROM identity.venues WHERE organization_id = $1", id)
			_, _ = pool.Exec(cleanup, "DELETE FROM identity.organization_members WHERE organization_id = $1", id)
			_, _ = pool.Exec(cleanup, "DELETE FROM identity.organizations WHERE id = $1", id)
		}
		_ = adminRepository.DeleteStaff(cleanup, admin.ID)
	}()

	// Capture contact details with the pending tenant and enforce the approval gate.
	created, err := tenantClient.CreateOrganization(ctx, &identityV1.TenantCreateOrganizationRequest{
		AccessToken: ownerToken, Name: "Registration test", Slug: "registration-" + uuid.NewString(), IdempotencyKey: uuid.NewString(),
		ContactName: "Venue contact", ContactEmail: "contact@example.test", ContactPhone: "0901234567", Note: "Please call in the morning",
	})
	if err != nil {
		t.Fatal(err)
	}
	id := created.GetOrganization().GetId()
	organizationIDs = append(organizationIDs, id)
	_, err = venueClient.CreateVenue(ctx, &identityV1.TenantCreateVenueRequest{AccessToken: ownerToken, OrganizationId: id, Name: "Venue", Slug: "venue-" + uuid.NewString()})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("pending tenant created venue: %v", err)
	}
	_, err = tenantClient.ListRegistrations(ctx, &identityV1.TenantListRegistrationsRequest{AccessToken: ownerToken})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("owner read admin contacts: %v", err)
	}
	registrations, err := tenantClient.ListRegistrations(ctx, &identityV1.TenantListRegistrationsRequest{AccessToken: adminToken, Status: "PENDING_VERIFICATION"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, registration := range registrations.GetItems() {
		if registration.GetOrganization().GetId() == id {
			found = registration.GetContactName() == "Venue contact" && registration.GetContactEmail() == "contact@example.test" && registration.GetContactPhone() == "0901234567"
		}
	}
	if !found {
		t.Fatal("contact details missing from admin queue")
	}
	decision := &identityV1.TenantReviewRegistrationRequest{AccessToken: adminToken, OrganizationId: id, Decision: "ACTIVE", Note: "Contact verified"}
	approved, err := tenantClient.ReviewRegistration(ctx, decision)
	if err != nil || approved.GetRegistration().GetOrganization().GetStatus() != "ACTIVE" {
		t.Fatalf("approve: %v %v", approved, err)
	}
	if _, err := tenantClient.ReviewRegistration(ctx, decision); err != nil {
		t.Fatalf("approval retry: %v", err)
	}
	var reviews int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM identity.tenant_registration_reviews WHERE organization_id=$1", id).Scan(&reviews); err != nil || reviews != 1 {
		t.Fatalf("audit count=%d err=%v", reviews, err)
	}
	_, err = tenantClient.ReviewRegistration(ctx, &identityV1.TenantReviewRegistrationRequest{AccessToken: adminToken, OrganizationId: id, Decision: "REJECTED"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("contradictory decision: %v", err)
	}

	// Create a venue only after approval and assign an eligible existing account.
	venue, err := venueClient.CreateVenue(ctx, &identityV1.TenantCreateVenueRequest{AccessToken: ownerToken, OrganizationId: id, Name: "Venue", Slug: "venue-" + uuid.NewString(), Address: "Test address"})
	if err != nil {
		t.Fatal(err)
	}
	member := tenantCreateTestUser(t, pool)
	_, err = staffClient.AddOrganizationStaff(ctx, &identityV1.AddOrganizationStaffRequest{AccessToken: ownerToken, OrganizationId: id, Email: member.Email, Role: constant.RoleVenueStaff})
	if err != nil {
		t.Fatal(err)
	}
	memberToken, _, err := tokens.NewAccessToken(member.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = staffClient.SetVenueStaff(ctx, &identityV1.TenantSetVenueStaffRequest{AccessToken: ownerToken, VenueId: venue.GetVenue().GetId(), UserId: member.ID, Status: "ACTIVE"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := venueClient.GetVenue(ctx, &identityV1.TenantGetVenueRequest{AccessToken: memberToken, VenueId: venue.GetVenue().GetId()}); err != nil {
		t.Fatalf("assigned venue: %v", err)
	}
	_, err = staffClient.ListOrganizationStaff(ctx, &identityV1.ListOrganizationStaffRequest{AccessToken: memberToken, OrganizationId: id})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("staff accessed owner-only directory: %v", err)
	}
	members, err := staffClient.ListOrganizationStaff(ctx, &identityV1.ListOrganizationStaffRequest{AccessToken: ownerToken, OrganizationId: id})
	if err != nil {
		t.Fatal(err)
	}
	found = false
	for _, entry := range members.GetItems() {
		if entry.GetUserId() == member.ID {
			found = len(entry.GetVenueIds()) == 1 && entry.GetEmail() == member.Email
		}
	}
	if !found {
		t.Fatal("staff directory lost venue assignment")
	}
	_, err = staffClient.RemoveOrganizationStaff(ctx, &identityV1.RemoveOrganizationStaffRequest{AccessToken: ownerToken, OrganizationId: id, UserId: member.ID})
	if err != nil {
		t.Fatal(err)
	}
	_, err = venueClient.GetVenue(ctx, &identityV1.TenantGetVenueRequest{AccessToken: memberToken, VenueId: venue.GetVenue().GetId()})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("revoked staff retained venue access: %v", err)
	}
}
