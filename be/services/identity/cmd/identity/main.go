package main

import (
	"context"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kainguyen/badminton-hub/be/platform/database"
	"github.com/kainguyen/badminton-hub/be/platform/grpc/server"
	"github.com/kainguyen/badminton-hub/be/platform/logger"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	authGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/auth"
	roleGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/role"
	staffGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/staff"
	tenantGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/tenant"
	userGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/user"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
)

func main() {
	logger := logger.New(platformService.Identity)
	if err := run(logger); err != nil {
		logger.Error("identity stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := identityConfig.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	pool, err := database.Open(ctx, cfg.DatabaseURL, cfg.DatabasePingTimeout)
	if err != nil {
		return err
	}
	defer pool.Close()

	userAuthServer := authGRPC.NewUserServer(cfg, pool)
	tenantAuthServer := authGRPC.NewTenantServer(cfg, pool)
	adminAuthServer := authGRPC.NewAdminServer(cfg, pool)
	staffServer := staffGRPC.NewStaffServer(cfg, pool)
	roleServer := roleGRPC.NewHandler(cfg, pool)
	userServer := userGRPC.NewUserServer(cfg, pool)
	tenantServer := tenantGRPC.NewServer(cfg, pool)
	return grpcServer.Run(
		ctx,
		cfg.GRPCAddr,
		func(server *grpc.Server) {
			identityV1.RegisterIdentityAuthServiceServer(server, userAuthServer)
			identityV1.RegisterIdentityTenantAuthServiceServer(server, tenantAuthServer)
			identityV1.RegisterIdentityAdminAuthServiceServer(server, adminAuthServer)
			identityV1.RegisterIdentityStaffServiceServer(server, staffServer)
			identityV1.RegisterIdentityRoleServiceServer(server, roleServer)
			identityV1.RegisterIdentityAdminServiceServer(server, &legacyAdminServer{
				AdminServer: adminAuthServer,
				StaffServer: staffServer,
				Handler:     roleServer,
			})
			identityV1.RegisterIdentityUserServiceServer(server, userServer)
			identityV1.RegisterIdentityTenantServiceServer(server, tenantServer)
		},
		cfg.ShutdownTimeout,
		logger,
	)
}

// legacyAdminServer preserves the previous gRPC endpoint during the gateway rollout.
type legacyAdminServer struct {
	legacyAdminFallback
	*authGRPC.AdminServer
	*staffGRPC.StaffServer
	*roleGRPC.Handler
}

type legacyAdminFallback struct {
	identityV1.UnimplementedIdentityAdminServiceServer
}
