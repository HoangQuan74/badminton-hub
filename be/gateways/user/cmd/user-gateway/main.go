package main

import (
	"context"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	"os"
	"os/signal"
	"syscall"

	userConfig "github.com/kainguyen/badminton-hub/be/gateways/user/config"
	userHTTP "github.com/kainguyen/badminton-hub/be/gateways/user/transport/http"
	"github.com/kainguyen/badminton-hub/be/platform/http/server"
	"github.com/kainguyen/badminton-hub/be/platform/logger"
)

func main() {
	logger := logger.New(platformService.UserGateway)
	cfg, err := userConfig.Load()
	if err != nil {
		logger.Error("user gateway configuration failed", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	engine, err := userHTTP.New(userHTTP.Dependencies{Config: cfg})
	if err != nil {
		logger.Error("user gateway initialization failed", "error", err)
		os.Exit(1)
	}
	if err := httpServer.Run(ctx, cfg.Address, engine, cfg.ShutdownTimeout, logger); err != nil {
		logger.Error("user gateway stopped", "error", err)
		os.Exit(1)
	}
}
