package app

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/platform/database"
	grpcServer "github.com/kainguyen/badminton-hub/be/platform/grpc/server"
	"github.com/kainguyen/badminton-hub/be/platform/logger"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
)

type Server struct {
	config identityConfig.Config
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewServer(ctx context.Context, cfg identityConfig.Config) (*Server, error) {
	// Open service-owned storage before accepting requests.
	pool, err := database.Open(
		ctx,
		cfg.DatabaseURL,
		cfg.DatabasePingTimeout,
	)
	if err != nil {
		return nil, err
	}

	return &Server{
		config: cfg,
		pool:   pool,
		logger: logger.New(platformService.Identity),
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	// Release storage after the transport has drained.
	defer s.pool.Close()
	return grpcServer.Run(
		ctx,
		s.config.GRPCAddr,
		RegisterGRPCServices(s.config, s.pool),
		s.config.ShutdownTimeout,
		s.logger,
	)
}
