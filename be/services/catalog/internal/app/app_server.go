package app

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/platform/database"
	grpcClient "github.com/kainguyen/badminton-hub/be/platform/grpc/client"
	grpcServer "github.com/kainguyen/badminton-hub/be/platform/grpc/server"
	"github.com/kainguyen/badminton-hub/be/platform/logger"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	catalogConfig "github.com/kainguyen/badminton-hub/be/services/catalog/config"
	identityAdapter "github.com/kainguyen/badminton-hub/be/services/catalog/internal/adapter/identity"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	catalogException "github.com/kainguyen/badminton-hub/be/services/catalog/internal/exception"
	catalogRepository "github.com/kainguyen/badminton-hub/be/services/catalog/internal/repository/postgres/catalog"
	s3store "github.com/kainguyen/badminton-hub/be/services/catalog/internal/storage/objectstore/s3"
	catalogGRPC "github.com/kainguyen/badminton-hub/be/services/catalog/internal/transport/grpc/catalog"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
)

type Server struct {
	config     catalogConfig.Config
	pool       *pgxpool.Pool
	connection *grpc.ClientConn
	transport  *catalogGRPC.CatalogServer
	logger     *slog.Logger
}

func NewServer(ctx context.Context, cfg catalogConfig.Config) (*Server, error) {
	pool, err := database.Open(
		ctx,
		cfg.DatabaseURL,
		cfg.DatabasePingTimeout,
	)
	if err != nil {
		return nil, err
	}

	// Create the Identity client used to verify each tenant request.
	identityConn, err := grpcClient.Dial(
		ctx,
		cfg.IdentityGRPCTarget,
		grpcClient.Options{ServiceIdentity: platformService.Catalog},
	)
	if err != nil {
		pool.Close()
		return nil, err
	}

	closeDependencies := func() {
		_ = identityConn.Close()
		pool.Close()
	}

	storage := catalogRepository.NewRepository(pool)
	ready, err := storage.IdentityCutoverReady(ctx)
	if err != nil {
		closeDependencies()
		return nil, err
	}

	if !ready {
		closeDependencies()
		return nil, catalogException.ErrIdentityCutoverRequired
	}

	directory := identityAdapter.NewDirectory(identityV1.NewIdentityVenueServiceClient(identityConn))
	repository := catalog.NewScopedRepository(storage, directory)
	service := catalog.NewService(repository)
	var images *catalog.ImageService
	if cfg.S3Enabled {
		objectStore, err := s3store.New(ctx, s3store.Config{
			Bucket:    cfg.S3Bucket,
			Region:    cfg.S3Region,
			Endpoint:  cfg.S3Endpoint,
			AccessKey: cfg.S3AccessKey,
			SecretKey: cfg.S3SecretKey,
		})
		if err != nil {
			closeDependencies()
			return nil, err
		}

		images = catalog.NewImageService(repository, objectStore)
	}

	transport := catalogGRPC.NewCatalogServer(
		service,
		identityV1.NewIdentityTenantAuthServiceClient(identityConn),
		images,
	)
	return &Server{
		config:     cfg,
		pool:       pool,
		connection: identityConn,
		transport:  transport,
		logger:     logger.New(platformService.Catalog),
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	defer s.pool.Close()
	defer s.connection.Close()
	return grpcServer.Run(
		ctx,
		s.config.GRPCAddr,
		func(server *grpc.Server) { catalogV1.RegisterCatalogServiceServer(server, s.transport) },
		s.config.ShutdownTimeout,
		s.logger,
	)
}
