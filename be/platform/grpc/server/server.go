// Package grpcServer owns the lifecycle and cross-cutting setup for internal
// gRPC servers. Generated service registration remains service-owned.
package grpcServer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/kainguyen/badminton-hub/be/platform/grpc/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Options struct {
	MaxRecvMsgSize     int
	MaxSendMsgSize     int
	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor
}

type Register func(*grpc.Server)

func Run(ctx context.Context, address string, register Register, shutdownTimeout time.Duration, logger *slog.Logger) error {
	return RunWithOptions(ctx, address, register, shutdownTimeout, Options{}, logger)
}

func RunWithOptions(ctx context.Context, address string, register Register, shutdownTimeout time.Duration, options Options, logger *slog.Logger) error {
	if shutdownTimeout <= 0 {
		shutdownTimeout = 10 * time.Second
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen gRPC: %w", err)
	}
	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(append([]grpc.UnaryServerInterceptor{grpcMetadata.UnaryServerInterceptor()}, options.UnaryInterceptors...)...),
		grpc.ChainStreamInterceptor(append([]grpc.StreamServerInterceptor{grpcMetadata.StreamServerInterceptor()}, options.StreamInterceptors...)...),
	}
	if options.MaxRecvMsgSize > 0 {
		serverOptions = append(serverOptions, grpc.MaxRecvMsgSize(options.MaxRecvMsgSize))
	}
	if options.MaxSendMsgSize > 0 {
		serverOptions = append(serverOptions, grpc.MaxSendMsgSize(options.MaxSendMsgSize))
	}
	server := grpc.NewServer(serverOptions...)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	if register != nil {
		register(server)
	}
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	serverErr := make(chan error, 1)
	go func() {
		logger.InfoContext(ctx, "grpc server started", "address", listener.Addr().String())
		serverErr <- server.Serve(listener)
	}()
	select {
	case err := <-serverErr:
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return fmt.Errorf("serve gRPC: %w", err)
	case <-ctx.Done():
		logger.InfoContext(ctx, "grpc shutdown signal received")
	}
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	stopped := make(chan struct{})
	go func() { server.GracefulStop(); close(stopped) }()
	select {
	case <-stopped:
	case <-shutdownCtx.Done():
		server.Stop()
		<-stopped
	}
	if err := <-serverErr; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("stop gRPC: %w", err)
	}
	logger.InfoContext(ctx, "grpc server stopped")
	return nil
}
