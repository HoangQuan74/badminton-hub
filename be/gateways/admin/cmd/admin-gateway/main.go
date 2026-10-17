package main

import (
	"context"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	"os"
	"os/signal"
	"syscall"

	adminConfig "github.com/kainguyen/badminton-hub/be/gateways/admin/config"
	adminHTTP "github.com/kainguyen/badminton-hub/be/gateways/admin/transport/http"
	"github.com/kainguyen/badminton-hub/be/platform/http/server"
	"github.com/kainguyen/badminton-hub/be/platform/logger"
)

func main() {
	log := logger.New(platformService.AdminGateway)
	cfg, err := adminConfig.Load()
	if err != nil {
		log.Error("admin gateway configuration failed", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	engine, err := adminHTTP.New(adminHTTP.Dependencies{Config: cfg})
	if err != nil {
		log.Error("admin gateway initialization failed", "error", err)
		os.Exit(1)
	}
	if err := httpServer.Run(ctx, cfg.Address, engine, cfg.ShutdownTimeout, log); err != nil {
		log.Error("admin gateway stopped", "error", err)
		os.Exit(1)
	}
}
