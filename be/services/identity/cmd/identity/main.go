package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/app"
)

func main() {
	if err := run(); err != nil {
		slog.Error("identity stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := identityConfig.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	server, err := app.NewServer(ctx, cfg)
	if err != nil {
		return err
	}
	return server.Run(ctx)
}
