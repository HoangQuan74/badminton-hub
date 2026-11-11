package main

import (
	"context"
	"net"
	"testing"

	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	authGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/auth"
	roleGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/role"
	staffGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/staff"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestLegacyAdminDispatch(t *testing.T) {
	cfg := identityConfig.Config{JWTSecret: "test-secret"}
	auth := authGRPC.NewAdminServer(cfg, nil)
	staff := staffGRPC.NewStaffServer(cfg, nil)
	role := roleGRPC.NewHandler(cfg, nil)
	server := grpc.NewServer()
	identityV1.RegisterIdentityAdminServiceServer(server, &legacyAdminServer{
		AdminServer: auth,
		StaffServer: staff,
		Handler:     role,
	})
	identityV1.RegisterIdentityAdminAuthServiceServer(server, auth)
	identityV1.RegisterIdentityStaffServiceServer(server, staff)
	identityV1.RegisterIdentityRoleServiceServer(server, role)
	listener := bufconn.Listen(1024 * 1024)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	conn, err := grpc.NewClient("passthrough:///buffer", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	legacy := identityV1.NewIdentityAdminServiceClient(conn)
	authClient := identityV1.NewIdentityAdminAuthServiceClient(conn)
	staffClient := identityV1.NewIdentityStaffServiceClient(conn)
	roleClient := identityV1.NewIdentityRoleServiceClient(conn)
	ctx := context.Background()
	tests := []struct {
		name    string
		legacy  func() error
		current func() error
	}{
		{"auth", func() error { _, err := legacy.Logout(ctx, &identityV1.AdminLogoutRequest{}); return err }, func() error { _, err := authClient.Logout(ctx, &identityV1.AdminLogoutRequest{}); return err }},
		{"staff", func() error { _, err := legacy.ListStaff(ctx, &identityV1.ListStaffRequest{}); return err }, func() error { _, err := staffClient.ListStaff(ctx, &identityV1.ListStaffRequest{}); return err }},
		{"role", func() error { _, err := legacy.ListRoles(ctx, &identityV1.ListRolesRequest{}); return err }, func() error { _, err := roleClient.ListRoles(ctx, &identityV1.ListRolesRequest{}); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oldCode, newCode := status.Code(test.legacy()), status.Code(test.current())
			if oldCode == codes.OK || oldCode == codes.Unimplemented || oldCode != newCode {
				t.Fatalf("legacy status %s; new status %s", oldCode, newCode)
			}
		})
	}
}
