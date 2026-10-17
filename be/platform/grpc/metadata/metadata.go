// Package grpcMetadata carries cross-cutting metadata across internal gRPC
// calls. It deliberately contains no service or business logic.
package grpcMetadata

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	RequestIDHeader     = "x-request-id"
	CorrelationIDHeader = "x-correlation-id"
	ServiceIdentityKey  = "x-service-identity"
	AuthorizationKey    = "authorization"
)

type contextKey struct{}

type RequestMetadata struct {
	RequestID       string
	CorrelationID   string
	ServiceIdentity string
}

func WithRequestMetadata(ctx context.Context, value RequestMetadata) context.Context {
	return context.WithValue(ctx, contextKey{}, value)
}

func FromContext(ctx context.Context) (RequestMetadata, bool) {
	value, ok := ctx.Value(contextKey{}).(RequestMetadata)
	return value, ok
}

// UnaryClientInterceptor propagates tracing identifiers and the caller's
// service identity. ServiceToken is intentionally supplied by the caller and
// is never logged or persisted by platform.
func UnaryClientInterceptor(serviceIdentity, serviceToken string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(appendOutgoing(ctx, serviceIdentity, serviceToken), method, request, reply, conn, opts...)
	}
}

func StreamClientInterceptor(serviceIdentity, serviceToken string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return streamer(appendOutgoing(ctx, serviceIdentity, serviceToken), desc, conn, method, opts...)
	}
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(contextWithIncoming(ctx), request)
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &wrappedServerStream{ServerStream: stream, ctx: contextWithIncoming(stream.Context())})
	}
}

func appendOutgoing(ctx context.Context, serviceIdentity, serviceToken string) context.Context {
	value, _ := FromContext(ctx)
	requestID := value.RequestID
	correlationID := value.CorrelationID
	if requestID == "" {
		requestID = uuid.NewString()
	}
	if correlationID == "" {
		correlationID = requestID
	}
	pairs := []string{RequestIDHeader, requestID, CorrelationIDHeader, correlationID}
	if serviceIdentity != "" {
		pairs = append(pairs, ServiceIdentityKey, serviceIdentity)
	}
	if serviceToken != "" {
		pairs = append(pairs, AuthorizationKey, "Bearer "+serviceToken)
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

func contextWithIncoming(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	requestID := validID(first(md.Get(RequestIDHeader)))
	if requestID == "" {
		requestID = uuid.NewString()
	}
	correlationID := validID(first(md.Get(CorrelationIDHeader)))
	if correlationID == "" {
		correlationID = requestID
	}
	return WithRequestMetadata(ctx, RequestMetadata{
		RequestID: requestID, CorrelationID: correlationID,
		ServiceIdentity: strings.TrimSpace(first(md.Get(ServiceIdentityKey))),
	})
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func validID(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > 128 {
		return ""
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || strings.ContainsRune("._:-", character)) {
			return ""
		}
	}
	return value
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *wrappedServerStream) Context() context.Context { return s.ctx }
