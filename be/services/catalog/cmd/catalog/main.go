package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kainguyen/badminton-hub/be/platform/database"
	grpcClient "github.com/kainguyen/badminton-hub/be/platform/grpc/client"
	grpcServer "github.com/kainguyen/badminton-hub/be/platform/grpc/server"
	"github.com/kainguyen/badminton-hub/be/platform/logger"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	catalogConfig "github.com/kainguyen/badminton-hub/be/services/catalog/config"
	identityAdapter "github.com/kainguyen/badminton-hub/be/services/catalog/internal/adapter/identity"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	s3store "github.com/kainguyen/badminton-hub/be/services/catalog/internal/storage/objectstore/s3"
	catalogRepository "github.com/kainguyen/badminton-hub/be/services/catalog/internal/storage/postgres/repository/catalog"
	catalogGRPC "github.com/kainguyen/badminton-hub/be/services/catalog/internal/transport/grpc/catalog"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
)

func main() {
	serviceLogger := logger.New(platformService.Catalog)
	if err := run(serviceLogger); err != nil {
		serviceLogger.Error("catalog stopped", "error", err)
		os.Exit(1)
	}
}

func run(serviceLogger *slog.Logger) error {
	cfg, err := catalogConfig.Load()
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

	// Create the Identity client used to verify each tenant request.
	identityConn, err := grpcClient.Dial(ctx, cfg.IdentityGRPCTarget, grpcClient.Options{ServiceIdentity: platformService.Catalog})
	if err != nil {
		return err
	}
	defer identityConn.Close()

	storage := catalogRepository.New(pool)
	ready, err := storage.IdentityCutoverReady(ctx)
	if err != nil {
		return err
	}
	if !ready {
		return errors.New("Identity tenant data transfer must be verified before Catalog startup")
	}
	directory := identityAdapter.New(identityV1.NewIdentityTenantServiceClient(identityConn))
	repository := catalog.NewScopedRepository(storage, directory)
	service := catalog.NewService(repository)
	var images *catalog.ImageService
	if cfg.S3Enabled {
		objectStore, err := s3store.New(ctx, s3store.Config{Bucket: cfg.S3Bucket, Region: cfg.S3Region, Endpoint: cfg.S3Endpoint, AccessKey: cfg.S3AccessKey, SecretKey: cfg.S3SecretKey})
		if err != nil {
			return err
		}
		images = catalog.NewImageService(repository, objectStore)
	}
	server := catalogGRPC.NewServer(service, identityV1.NewIdentityAuthServiceClient(identityConn), images)
	return grpcServer.Run(ctx, cfg.GRPCAddr, func(grpcServer *grpc.Server) {
		catalogV1.RegisterCatalogServiceServer(grpcServer, server)
	}, cfg.ShutdownTimeout, serviceLogger)
}
