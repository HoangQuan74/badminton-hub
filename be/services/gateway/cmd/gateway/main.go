package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kainguyen/badminton-hub/be/platform/httpserver"
	gatewayconfig "github.com/kainguyen/badminton-hub/be/services/gateway/config"
	gatewayhttp "github.com/kainguyen/badminton-hub/be/services/gateway/transport/http"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "gateway")
	if err := run(logger); err != nil {
		logger.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := gatewayconfig.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	engine, err := gatewayhttp.New(gatewayhttp.Dependencies{Config: cfg})
	if err != nil {
		return err
	}
	return httpserver.Run(ctx, cfg.HTTPAddr, engine, cfg.ShutdownTimeout, logger)
}
