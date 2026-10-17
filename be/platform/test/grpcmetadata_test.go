package test

import (
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	"context"
	"testing"

	grpcMetadata "github.com/kainguyen/badminton-hub/be/platform/grpc/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryInterceptorsPropagateAndReadMetadata(t *testing.T) {
	ctx := grpcMetadata.WithRequestMetadata(context.Background(), grpcMetadata.RequestMetadata{
		RequestID: "request-123", CorrelationID: "correlation-456",
	})
	client := grpcMetadata.UnaryClientInterceptor(platformService.Identity, "secret")
	var outgoing metadata.MD
	err := client(ctx, "/identity.v1.Auth/Me", nil, nil, nil, func(callCtx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		outgoing, _ = metadata.FromOutgoingContext(callCtx)
		return nil
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := outgoing.Get(grpcMetadata.RequestIDHeader); len(got) != 1 || got[0] != "request-123" {
		t.Fatalf("request metadata = %#v", got)
	}
	if got := outgoing.Get(grpcMetadata.AuthorizationKey); len(got) != 1 || got[0] != "Bearer secret" {
		t.Fatalf("authorization metadata = %#v", got)
	}

	incoming := metadata.NewIncomingContext(context.Background(), outgoing)
	_, err = grpcMetadata.UnaryServerInterceptor()(incoming, nil, &grpc.UnaryServerInfo{}, func(handlerCtx context.Context, _ any) (any, error) {
		got, ok := grpcMetadata.FromContext(handlerCtx)
		if !ok || got.RequestID != "request-123" || got.CorrelationID != "correlation-456" || got.ServiceIdentity != platformService.Identity {
			t.Fatalf("server metadata = %#v, present = %v", got, ok)
		}
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServerGeneratesMissingIdentifiers(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(grpcMetadata.ServiceIdentityKey, "catalog"))
	_, err := grpcMetadata.UnaryServerInterceptor()(ctx, nil, &grpc.UnaryServerInfo{}, func(handlerCtx context.Context, _ any) (any, error) {
		got, ok := grpcMetadata.FromContext(handlerCtx)
		if !ok || got.RequestID == "" || got.CorrelationID == "" || got.RequestID != got.CorrelationID || got.ServiceIdentity != "catalog" {
			t.Fatalf("generated metadata = %#v, present = %v", got, ok)
		}
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
