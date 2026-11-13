package identitytest

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	identityApp "github.com/kainguyen/badminton-hub/be/services/identity/internal/app"
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

func TestPermissionGRPCCRUDAuthorization(t *testing.T) {
	ctx := context.Background()
	pool := staffTestPool(t)
	repository := staffPostgres.NewRepository(pool)
	admin, err := repository.CreateStaff(ctx, staff.CreateStaffParams{Email: uuid.NewString() + "@example.test", PasswordHash: "unused", FullName: "Permission admin", IsAdmin: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repository.DeleteStaff(context.Background(), admin.ID) })
	cfg := identityConfig.Config{JWTSecret: "permission-integration-only-secret-long-enough", AccessTokenTTL: time.Minute, RefreshTokenTTL: time.Hour}
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	access, _, err := tokens.NewAccessToken(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	identityApp.RegisterGRPCServices(cfg, pool)(server)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	connection, err := grpc.NewClient("passthrough:///permission-test", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	client := identityV1.NewIdentityPermissionServiceClient(connection)
	if _, err := client.ListPermissions(ctx, &identityV1.ListPermissionsRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("missing token: %v", err)
	}
	created, err := client.CreatePermission(ctx, &identityV1.CreatePermissionRequest{AccessToken: access, Module: " Tenant ", Path: "/test/" + uuid.NewString(), Action: "read"})
	if err != nil {
		t.Fatal(err)
	}
	id := created.GetPermission().GetId()
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM identity.permissions WHERE id = $1", id) })
	if created.GetPermission().GetModule() != "tenant" || created.GetPermission().GetAction() != "READ" {
		t.Fatalf("canonical values: %v", created)
	}
	_, err = client.CreatePermission(ctx, &identityV1.CreatePermissionRequest{AccessToken: access, Module: "TENANT", Path: created.GetPermission().GetPath(), Action: "READ"})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate: %v", err)
	}
	got, err := client.GetPermission(ctx, &identityV1.GetPermissionRequest{AccessToken: access, PermissionId: id})
	if err != nil || got.GetPermission().GetId() != id {
		t.Fatalf("get: %v %v", got, err)
	}
	updated, err := client.UpdatePermission(ctx, &identityV1.UpdatePermissionRequest{AccessToken: access, PermissionId: id, Module: "tenant", Path: created.GetPermission().GetPath(), Action: "UPDATE", Description: "updated"})
	if err != nil || updated.GetPermission().GetDescription() != "updated" {
		t.Fatalf("update: %v %v", updated, err)
	}
	if _, err := client.DeletePermission(ctx, &identityV1.DeletePermissionRequest{AccessToken: access, PermissionId: id}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetPermission(ctx, &identityV1.GetPermissionRequest{AccessToken: access, PermissionId: id}); status.Code(err) != codes.NotFound {
		t.Fatalf("deleted permission: %v", err)
	}
	if _, err := client.GetPermission(ctx, &identityV1.GetPermissionRequest{AccessToken: access, PermissionId: "invalid"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid ID: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE identity.users SET status = 'suspended' WHERE id = $1", admin.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListPermissions(ctx, &identityV1.ListPermissionsRequest{AccessToken: access}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("suspended admin: %v", err)
	}
}
