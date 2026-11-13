package identitytest

import (
	"context"
	"net"
	"testing"

	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	identityApp "github.com/kainguyen/badminton-hub/be/services/identity/internal/app"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestAppRegistersModuleServices(t *testing.T) {
	cfg := identityConfig.Config{JWTSecret: "test-secret"}
	server := grpc.NewServer()
	identityApp.RegisterGRPCServices(cfg, nil)(server)
	listener := bufconn.Listen(1024 * 1024)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	conn, err := grpc.NewClient("passthrough:///buffer", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	authClient := identityV1.NewIdentityAdminAuthServiceClient(conn)
	staffClient := identityV1.NewIdentityStaffServiceClient(conn)
	roleClient := identityV1.NewIdentityRoleServiceClient(conn)
	permissionClient := identityV1.NewIdentityPermissionServiceClient(conn)
	tenantClient := identityV1.NewIdentityTenantServiceClient(conn)
	venueClient := identityV1.NewIdentityVenueServiceClient(conn)
	ctx := context.Background()
	tests := []struct {
		name string
		call func() error
	}{
		{"auth", func() error { _, err := authClient.Logout(ctx, &identityV1.AdminLogoutRequest{}); return err }},
		{"staff", func() error { _, err := staffClient.ListStaff(ctx, &identityV1.ListStaffRequest{}); return err }},
		{"role", func() error { _, err := roleClient.ListRoles(ctx, &identityV1.ListRolesRequest{}); return err }},
		{"permission", func() error {
			_, err := permissionClient.ListPermissions(ctx, &identityV1.ListPermissionsRequest{})
			return err
		}},
		{"tenant", func() error {
			_, err := tenantClient.ListOrganizations(ctx, &identityV1.TenantListOrganizationsRequest{})
			return err
		}},
		{"venue", func() error { _, err := venueClient.ListVenues(ctx, &identityV1.TenantListVenuesRequest{}); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if code := status.Code(test.call()); code != codes.Unauthenticated {
				t.Fatalf("expected authentication guard, got %s", code)
			}
		})
	}
}
