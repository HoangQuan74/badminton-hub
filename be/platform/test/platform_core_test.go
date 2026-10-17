package test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	platformConfig "github.com/kainguyen/badminton-hub/be/platform/config"
	grpcClient "github.com/kainguyen/badminton-hub/be/platform/grpc/client"
	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	grpcMetadata "github.com/kainguyen/badminton-hub/be/platform/grpc/metadata"
	grpcServer "github.com/kainguyen/badminton-hub/be/platform/grpc/server"
	httpServer "github.com/kainguyen/badminton-hub/be/platform/http/server"
	platformLogger "github.com/kainguyen/badminton-hub/be/platform/logger"
	platformMetrics "github.com/kainguyen/badminton-hub/be/platform/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestConfigHelpers(t *testing.T) {
	t.Setenv("PLATFORM_TEST_VALUE", "configured")
	if got := platformConfig.Env("PLATFORM_TEST_VALUE", "fallback"); got != "configured" {
		t.Fatalf("Env() = %q", got)
	}
	if got := platformConfig.Env("PLATFORM_MISSING_VALUE", "fallback"); got != "fallback" {
		t.Fatalf("Env() = %q", got)
	}
	if got, err := platformConfig.Required("PLATFORM_TEST_VALUE"); err != nil || got != "configured" {
		t.Fatalf("Required() = %q, %v", got, err)
	}
	if _, err := platformConfig.Required("PLATFORM_MISSING_VALUE"); err == nil {
		t.Fatal("Required() accepted missing value")
	}
	if got, err := platformConfig.Duration("PLATFORM_DURATION", time.Second); err != nil || got != time.Second {
		t.Fatalf("Duration fallback = %v, %v", got, err)
	}
	t.Setenv("PLATFORM_DURATION", "250ms")
	if got, err := platformConfig.Duration("PLATFORM_DURATION", time.Second); err != nil || got != 250*time.Millisecond {
		t.Fatalf("Duration configured = %v, %v", got, err)
	}
	for _, raw := range []string{"0s", "invalid"} {
		t.Setenv("PLATFORM_DURATION", raw)
		if _, err := platformConfig.Duration("PLATFORM_DURATION", time.Second); err == nil {
			t.Fatalf("Duration accepted %q", raw)
		}
	}
}

func TestGRPCEndpointAuthorizationAndExecutionErrors(t *testing.T) {
	mapper := grpcError.NewErrorMapper(codes.Internal, "internal")
	if _, err := grpcEndpoint.Run(context.Background(), "request", mapper, nil, func(context.Context, string) (string, error) { return "ok", nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := grpcEndpoint.Run(context.Background(), "request", mapper, func(context.Context, string) error { return errors.New("denied") }, func(context.Context, string) (string, error) { return "", nil }); status.Code(err) != codes.Internal {
		t.Fatalf("authorization code = %v", status.Code(err))
	}
	if _, err := grpcEndpoint.Run(context.Background(), "request", mapper, nil, func(context.Context, string) (string, error) { return "", errors.New("failed") }); status.Code(err) != codes.Internal {
		t.Fatalf("execution code = %v", status.Code(err))
	}
	authorize := grpcEndpoint.Authorize[accessTokenRequest](func(_ context.Context, token string) error {
		if token != "token" {
			return errors.New("bad token")
		}
		return nil
	})
	if err := authorize(context.Background(), accessTokenRequest{"token"}); err != nil {
		t.Fatal(err)
	}
	if err := authorize(context.Background(), accessTokenRequest{"wrong"}); err == nil {
		t.Fatal("Authorize accepted invalid token")
	}
}

type accessTokenRequest struct{ token string }

func (r accessTokenRequest) GetAccessToken() string { return r.token }

func TestGRPCErrorMapperNilAndWrappedRule(t *testing.T) {
	target := errors.New("target")
	mapper := grpcError.NewErrorMapper(codes.Internal, "fallback", grpcError.MatchError(target, codes.InvalidArgument, "invalid"))
	if mapper.Map(nil) != nil {
		t.Fatal("Map(nil) returned an error")
	}
	if status.Code(mapper.Map(errors.Join(errors.New("context"), target))) != codes.InvalidArgument {
		t.Fatal("wrapped target did not match")
	}
}

func TestGRPCMetadataStreamInterceptorsAndInvalidIDs(t *testing.T) {
	ctx := grpcMetadata.WithRequestMetadata(context.Background(), grpcMetadata.RequestMetadata{RequestID: "req", CorrelationID: "corr"})
	client := grpcMetadata.StreamClientInterceptor("svc", "token")
	_, err := client(ctx, &grpc.StreamDesc{}, nil, "/method", func(callCtx context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
		md, _ := metadata.FromOutgoingContext(callCtx)
		if md.Get(grpcMetadata.AuthorizationKey)[0] != "Bearer token" {
			t.Fatal("stream token not propagated")
		}
		return nil, nil
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	stream := &testServerStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(grpcMetadata.RequestIDHeader, "bad id"))}
	if err := grpcMetadata.StreamServerInterceptor()(nil, stream, &grpc.StreamServerInfo{}, func(_ any, got grpc.ServerStream) error {
		value, ok := grpcMetadata.FromContext(got.Context())
		if !ok || value.RequestID == "bad id" || value.RequestID == "" {
			t.Fatal("invalid request ID was accepted")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	client = grpcMetadata.StreamClientInterceptor("", "")
	if _, err := client(context.Background(), &grpc.StreamDesc{}, nil, "/method", func(callCtx context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
		md, _ := metadata.FromOutgoingContext(callCtx)
		if len(md.Get(grpcMetadata.AuthorizationKey)) != 0 {
			t.Fatal("unexpected authorization metadata")
		}
		return nil, nil
	}, nil); err != nil {
		t.Fatal(err)
	}
}

type testServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *testServerStream) Context() context.Context { return s.ctx }

func TestGRPCClientDialValidationAndNonBlockingConnection(t *testing.T) {
	if _, err := grpcClient.Dial(context.Background(), "", grpcClient.Options{}); err == nil {
		t.Fatal("Dial accepted empty target")
	}
	conn, err := grpcClient.Dial(context.Background(), "passthrough:///unused", grpcClient.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := grpcClient.Dial(context.Background(), "passthrough:///unused", grpcClient.Options{Block: true, DialTimeout: time.Millisecond}); err == nil {
		t.Fatal("blocking Dial unexpectedly succeeded")
	}
}

func TestServerLifecycles(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(10 * time.Millisecond); cancel() }()
	if err := grpcServer.Run(ctx, "127.0.0.1:0", nil, time.Second, logger); err != nil {
		t.Fatal(err)
	}
	if err := grpcServer.Run(context.Background(), "bad address", nil, time.Second, logger); err == nil {
		t.Fatal("invalid gRPC address accepted")
	}
	ctx, cancel = context.WithCancel(context.Background())
	go func() { time.Sleep(10 * time.Millisecond); cancel() }()
	if err := grpcServer.RunWithOptions(ctx, "127.0.0.1:0", nil, 0, grpcServer.Options{
		MaxRecvMsgSize: 1024,
		MaxSendMsgSize: 1024,
		UnaryInterceptors: []grpc.UnaryServerInterceptor{func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, request)
		}},
		StreamInterceptors: []grpc.StreamServerInterceptor{func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return handler(srv, stream)
		}},
	}, logger); err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	go func() { time.Sleep(10 * time.Millisecond); cancel() }()
	if err := httpServer.Run(ctx, "127.0.0.1:0", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), time.Second, logger); err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	go func() { time.Sleep(10 * time.Millisecond); cancel() }()
	if err := httpServer.RunWithOptions(ctx, "127.0.0.1:0", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), 0, httpServer.Options{}, logger); err != nil {
		t.Fatal(err)
	}
}

func TestLoggerConstructorsAndGroups(t *testing.T) {
	var output strings.Builder
	logger := platformLogger.New("service")
	if logger == nil {
		t.Fatal("New returned nil")
	}
	configured := platformLogger.NewWithWriter("service", &output).WithGroup("request")
	configured.Info("message")
	if !strings.Contains(output.String(), "message") {
		t.Fatal("grouped logger did not write")
	}
}

func TestMetricsEmptyAndUnknownRoute(t *testing.T) {
	registry := platformMetrics.New("service")
	engine := httptest.NewRecorder()
	registry.Handler().ServeHTTP(engine, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(engine.Body.String(), "http_in_flight_requests") {
		t.Fatal("empty metrics missing gauge")
	}
	ginEngine := newTestGinEngine(registry)
	response := httptest.NewRecorder()
	ginEngine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/unknown", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("unknown route status = %d", response.Code)
	}
	ginContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	registry.Middleware()(ginContext)
}

func newTestGinEngine(registry *platformMetrics.Registry) *gin.Engine {
	engine := gin.New()
	engine.Use(registry.Middleware())
	return engine
}
