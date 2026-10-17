package test

import (
	"context"
	"errors"
	"testing"

	"github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGRPCEndpointRunsAuthorizationAndOperation(t *testing.T) {
	mapper := grpcError.NewErrorMapper(codes.Internal, "internal error")
	result, err := grpcEndpoint.Run(context.Background(), "request", mapper,
		func(context.Context, string) error { return nil },
		func(context.Context, string) (string, error) { return "response", nil },
	)
	if err != nil || result != "response" {
		t.Fatalf("result = %q, error = %v", result, err)
	}
}

func TestGRPCEndpointMapsAuthorizationError(t *testing.T) {
	mapper := grpcError.NewErrorMapper(codes.Internal, "internal error")
	result, err := grpcEndpoint.Run(context.Background(), "request", mapper,
		func(context.Context, string) error { return errors.New("denied") },
		func(context.Context, string) (string, error) { return "unexpected", nil },
	)
	if result != "" || status.Code(err) != codes.Internal {
		t.Fatalf("result = %q, error = %v", result, err)
	}
}
