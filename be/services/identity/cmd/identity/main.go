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
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/auth"
	staffGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/staff"
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

	authServer := identityGRPC.NewAuthServer(cfg, pool)
	staffServer := staffGRPC.NewStaffServer(cfg, pool)
	userServer := userGRPC.NewUserServer(cfg, pool)
	return grpcServer.Run(
		ctx,
		cfg.GRPCAddr,
		func(server *grpc.Server) {
			identityV1.RegisterIdentityAuthServiceServer(server, authServer)
			identityV1.RegisterIdentityAdminServiceServer(server, staffServer)
			identityV1.RegisterIdentityUserServiceServer(server, userServer)
		},
		cfg.ShutdownTimeout,
		logger,
	)
}
