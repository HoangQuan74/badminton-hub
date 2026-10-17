package main

import (
	"context"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	grpcServer "github.com/kainguyen/badminton-hub/be/platform/grpc/server"
	"github.com/kainguyen/badminton-hub/be/platform/logger"
	socketConfig "github.com/kainguyen/badminton-hub/be/services/socket/config"
	realtimeGRPC "github.com/kainguyen/badminton-hub/be/services/socket/internal/transport/grpc/realtime"
	socketV1 "github.com/kainguyen/badminton-hub/contracts/gen/socket/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	log := logger.New(platformService.Socket)
	if err := run(log); err != nil {
		log.Error("socket stopped", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := socketConfig.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return err
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return err
	}
	server := realtimeGRPC.NewServer(redisClient, cfg.RedisChannel, cfg.JWTSecret)
	return grpcServer.Run(ctx, cfg.GRPCAddr, func(grpcServer *grpc.Server) {
		socketV1.RegisterSocketServiceServer(grpcServer, server)
	}, cfg.ShutdownTimeout, log)
}
