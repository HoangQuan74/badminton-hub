package test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/kainguyen/badminton-hub/be/platform/http/context"
	platformLogger "github.com/kainguyen/badminton-hub/be/platform/logger"
	"go.opentelemetry.io/otel/trace"
)

func TestNewWithWriterAddsOpenTelemetrySpanFields(t *testing.T) {
	var output bytes.Buffer
	logger := platformLogger.NewWithWriter("test-service", &output)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1, 2, 3},
		SpanID:     trace.SpanID{4, 5, 6},
		TraceFlags: trace.FlagsSampled,
	})
	logger.InfoContext(trace.ContextWithSpanContext(context.Background(), spanContext), "started")

	line := output.String()
	for _, field := range []string{`"service":"test-service"`, `"trace_id":"01020300000000000000000000000000"`, `"span_id":"0405060000000000"`, `"trace_sampled":true`} {
		if !strings.Contains(line, field) {
			t.Fatalf("log missing %s: %s", field, line)
		}
	}
}

func TestTraceHandlerPreservesWithAttrs(t *testing.T) {
	var output bytes.Buffer
	logger := platformLogger.NewWithWriter("test-service", &output).With("operation", "startup")
	logger.Log(context.Background(), slog.LevelInfo, "ready")
	if !strings.Contains(output.String(), `"operation":"startup"`) {
		t.Fatalf("log missing operation: %s", output.String())
	}
}

func TestLoggerAddsRequestMetadata(t *testing.T) {
	var output bytes.Buffer
	logger := platformLogger.NewWithWriter("test-service", &output)
	ctx := httpContext.WithRequestMetadata(context.Background(), httpContext.RequestMetadata{
		RequestID: "request-123", CorrelationID: "correlation-456",
	})
	logger.InfoContext(ctx, "request completed")
	line := output.String()
	for _, field := range []string{`"request_id":"request-123"`, `"correlation_id":"correlation-456"`} {
		if !strings.Contains(line, field) {
			t.Fatalf("log missing %s: %s", field, line)
		}
	}
}
