package main

import (
	"context"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	socketConfig "github.com/kainguyen/badminton-hub/be/gateways/socket/config"
	socketHTTP "github.com/kainguyen/badminton-hub/be/gateways/socket/transport/http"
	"github.com/kainguyen/badminton-hub/be/platform/http/server"
	"github.com/kainguyen/badminton-hub/be/platform/logger"
)

func main() {
	logger := logger.New(platformService.SocketGateway)
	cfg, err := socketConfig.Load()
	if err != nil {
		logger.Error("socket gateway configuration failed", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	engine, err := socketHTTP.New(socketHTTP.Dependencies{Config: cfg})
	if err != nil {
		logger.Error("socket gateway initialization failed", "error", err)
		os.Exit(1)
	}
	if err := httpServer.RunWithOptions(ctx, cfg.Address, engine, cfg.ShutdownTimeout, httpServer.Options{
		ReadHeaderTimeout: 5 * time.Second,
	}, logger); err != nil {
		logger.Error("socket gateway stopped", "error", err)
		os.Exit(1)
	}
}
