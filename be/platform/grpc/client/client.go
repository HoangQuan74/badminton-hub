// Package grpcClient centralizes internal gRPC dialing and metadata
// propagation. Generated clients remain owned by their consuming service.
package grpcClient

import (
	"context"
	"fmt"
	"time"

	"github.com/kainguyen/badminton-hub/be/platform/grpc/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Options struct {
	DialTimeout     time.Duration
	ServiceIdentity string
	ServiceToken    string
	Credentials     credentials.TransportCredentials
	Block           bool
}

func Dial(ctx context.Context, target string, options Options) (*grpc.ClientConn, error) {
	if target == "" {
		return nil, fmt.Errorf("gRPC target is required")
	}
	if options.DialTimeout <= 0 {
		options.DialTimeout = 5 * time.Second
	}
	dialCtx, cancel := context.WithTimeout(ctx, options.DialTimeout)
	defer cancel()
	transportCredentials := options.Credentials
	if transportCredentials == nil {
		transportCredentials = insecure.NewCredentials()
	}
	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCredentials),
		grpc.WithUnaryInterceptor(grpcMetadata.UnaryClientInterceptor(options.ServiceIdentity, options.ServiceToken)),
		grpc.WithStreamInterceptor(grpcMetadata.StreamClientInterceptor(options.ServiceIdentity, options.ServiceToken)),
	}
	if options.Block {
		dialOptions = append(dialOptions, grpc.WithBlock())
	}
	conn, err := grpc.DialContext(dialCtx, target, dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("dial gRPC %q: %w", target, err)
	}
	return conn, nil
}
