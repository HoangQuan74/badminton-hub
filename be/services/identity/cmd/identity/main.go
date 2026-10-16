package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kainguyen/badminton-hub/be/platform/database"
	"github.com/kainguyen/badminton-hub/be/platform/httpserver"
	identityconfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	identityhttp "github.com/kainguyen/badminton-hub/be/services/identity/transport/http"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "identity")
	if err := run(logger); err != nil {
		logger.Error("identity stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := identityconfig.Load()
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

	engine, err := identityhttp.New(identityhttp.Dependencies{Config: cfg, Pool: pool, Logger: logger})
	if err != nil {
		return err
	}
	return httpserver.Run(ctx, cfg.HTTPAddr, engine, cfg.ShutdownTimeout, logger)
}
