package identitytest

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/role"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/security"
	authpostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/auth"
	rolepostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/role"
	staffpostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/staff"
	userpostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/repository/user"
	authgrpc "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/auth"
	rolegrpc "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/role"
	staffgrpc "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/staff"
	usergrpc "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/user"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// These tests write to the supplied database. Use a dedicated disposable database.
func staffTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_IDENTITY_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_IDENTITY_DATABASE_URL must point to a disposable migrated database")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestAdminStaffGRPCContract(t *testing.T) {
	ctx := context.Background()
	pool := staffTestPool(t)
	repository := staffpostgres.NewRepository(pool)
	hash, err := security.HashPassword("password-one")
	if err != nil {
		t.Fatal(err)
	}
	admin, err := repository.Create(ctx, staff.CreateParams{Email: uuid.NewString() + "@example.test", PasswordHash: hash, FullName: "Admin API test", IsAdmin: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repository.Delete(context.Background(), admin.ID) })
	cfg := identityConfig.Config{JWTSecret: "integration-test-secret-with-sufficient-length", AccessTokenTTL: time.Minute, RefreshTokenTTL: time.Hour}
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	identityV1.RegisterIdentityAdminAuthServiceServer(grpcServer, authgrpc.NewAdminServer(cfg, pool))
	identityV1.RegisterIdentityStaffServiceServer(grpcServer, staffgrpc.NewStaffServer(cfg, pool))
	identityV1.RegisterIdentityRoleServiceServer(grpcServer, rolegrpc.NewHandler(cfg, pool))
	identityV1.RegisterIdentityAuthServiceServer(grpcServer, authgrpc.NewUserServer(cfg, pool))
	identityV1.RegisterIdentityUserServiceServer(grpcServer, usergrpc.NewUserServer(cfg, pool))
	go func() { _ = grpcServer.Serve(listener) }()
	t.Cleanup(grpcServer.Stop)
	conn, err := grpc.NewClient("passthrough:///buffer", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	server := identityV1.NewIdentityAdminAuthServiceClient(conn)
	staffClient := identityV1.NewIdentityStaffServiceClient(conn)
	roleClient := identityV1.NewIdentityRoleServiceClient(conn)
	login, err := server.Login(ctx, &identityV1.AdminLoginRequest{Email: admin.Email, Password: "password-one"})
	if err != nil {
		t.Fatal(err)
	}
	access := login.GetTokens().GetAccessToken()
	if _, err := staffClient.ListStaff(ctx, &identityV1.ListStaffRequest{AccessToken: access, Limit: 10}); err != nil {
		t.Fatal(err)
	}
	userClient := identityV1.NewIdentityUserServiceClient(conn)
	userEmail := uuid.NewString() + "@example.test"
	createdUser, err := userClient.CreateUser(ctx, &identityV1.CreateUserRequest{AccessToken: access, Email: userEmail, Password: "user-password", FullName: "User API test"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = userpostgres.NewRepository(pool).Delete(context.Background(), createdUser.GetId()) })
	authClient := identityV1.NewIdentityAuthServiceClient(conn)
	if _, err := authClient.Login(ctx, &identityV1.LoginRequest{Email: userEmail, Password: "user-password", Portal: "user"}); err != nil {
		t.Fatalf("user login: %v", err)
	}
	created, err := roleClient.CreateRole(ctx, &identityV1.CreateRoleRequest{AccessToken: access, Name: "grpc-" + uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = rolepostgres.NewRepository(pool).DeleteRole(context.Background(), created.GetRole().GetId())
	})
	if created.GetRole().GetCreatedAt() == "" {
		t.Fatal("missing role creation time")
	}
	if _, err := roleClient.ListRoles(ctx, &identityV1.ListRolesRequest{AccessToken: access}); err != nil {
		t.Fatal(err)
	}
	if _, err := roleClient.ListPermissions(ctx, &identityV1.ListPermissionsRequest{AccessToken: access}); err != nil {
		t.Fatal(err)
	}
	if _, err := roleClient.ListRoles(ctx, &identityV1.ListRolesRequest{AccessToken: "invalid"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("invalid bearer: %v", err)
	}
	refresh, err := server.Refresh(ctx, &identityV1.AdminRefreshRequest{RefreshToken: login.GetTokens().GetRefreshToken()})
	if err != nil || refresh.GetTokens().GetRefreshToken() == "" {
		t.Fatalf("refresh: %v", err)
	}
	if _, err := server.Refresh(ctx, &identityV1.AdminRefreshRequest{RefreshToken: login.GetTokens().GetRefreshToken()}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("replay: %v", err)
	}
}

func TestAdminStaffSessionsAndPassword(t *testing.T) {
	ctx := context.Background()
	pool := staffTestPool(t)
	repository := staffpostgres.NewRepository(pool)
	hash, err := security.HashPassword("password-one")
	if err != nil {
		t.Fatal(err)
	}
	created, err := repository.Create(ctx, staff.CreateParams{Email: uuid.NewString() + "@example.test", PasswordHash: hash, FullName: "Admin test", IsAdmin: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repository.Delete(context.Background(), created.ID) })
	service := auth.NewAdminService(authpostgres.NewAuthRepository(pool), auth.NewTokenManager("integration-test-secret-with-sufficient-length", time.Minute, time.Hour))

	if _, _, err := service.Login(ctx, created.Email, "wrong"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("wrong password: %v", err)
	}
	_, pair, err := service.Login(ctx, created.Email, "password-one")
	if err != nil {
		t.Fatal(err)
	}

	// Two refreshes of the same token must produce only one replacement session.
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := service.Refresh(ctx, pair.RefreshToken); results <- err }()
	}
	wg.Wait()
	close(results)
	successes, rejected := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, auth.ErrInvalidToken) {
			rejected++
		} else {
			t.Fatalf("refresh: %v", err)
		}
	}
	if successes != 1 || rejected != 1 {
		t.Fatalf("refresh results success=%d rejected=%d", successes, rejected)
	}
	if err := service.ChangePassword(ctx, created.ID, "wrong", "password-two"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("incorrect current password: %v", err)
	}
	if err := service.ChangePassword(ctx, created.ID, "password-one", "password-two"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.Login(ctx, created.Email, "password-two"); err != nil {
		t.Fatalf("new password: %v", err)
	}
	if _, err := service.Refresh(ctx, pair.RefreshToken); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("replayed refresh: %v", err)
	}
	if err := service.Logout(ctx, ""); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("empty logout: %v", err)
	}
}

func TestStaffRolePermissionsRollback(t *testing.T) {
	ctx := context.Background()
	pool := staffTestPool(t)
	repository := rolepostgres.NewRepository(pool)
	createdRole, err := repository.CreateRole(ctx, "integration-"+uuid.NewString(), "original")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repository.DeleteRole(context.Background(), createdRole.ID) })
	permissionID := uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO identity.permissions (id, module, path, action) VALUES ($1, $2, $3, 'READ')`, permissionID, "test", uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.permissions WHERE id = $1`, permissionID)
	})
	if _, err := repository.ReplaceRolePermissions(ctx, createdRole.ID, []string{permissionID.String()}); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.ReplaceRolePermissions(ctx, createdRole.ID, []string{uuid.NewString()}); err == nil {
		t.Fatal("unknown permission unexpectedly accepted")
	}
	var count int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM identity.role_permissions WHERE role_id = $1 AND permission_id = $2`, createdRole.ID, permissionID).Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("replacement rollback: count=%d err=%v", count, err)
	}
	updated, err := repository.UpdateRole(ctx, createdRole.ID, createdRole.Name, "updated")
	if err != nil || updated.Description != "updated" {
		t.Fatalf("update role: %+v %v", updated, err)
	}
	if _, err := repository.GetRole(ctx, uuid.NewString()); !errors.Is(err, role.ErrNotFound) {
		t.Fatalf("missing role: %v", err)
	}
	if _, err := repository.GetRole(ctx, "bad-id"); !errors.Is(err, role.ErrInvalidInput) {
		t.Fatalf("invalid role: %v", err)
	}
}
