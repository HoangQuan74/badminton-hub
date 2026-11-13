package main

import (
	"context"
	catalogConfig "github.com/kainguyen/badminton-hub/be/services/catalog/config"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/app"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		slog.Error("catalog stopped", "error", err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := catalogConfig.Load()
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
