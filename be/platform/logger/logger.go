// Package logger provides the process-wide structured logger.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/kainguyen/badminton-hub/be/platform/http/context"
	"go.opentelemetry.io/otel/trace"
)

// New creates a JSON logger and correlates context-aware records with the
// active OpenTelemetry span.
func New(service string) *slog.Logger {
	return NewWithWriter(service, os.Stdout)
}

// NewWithWriter is configurable for tests and callers that own process output.
func NewWithWriter(service string, writer io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{AddSource: true})
	return slog.New(traceHandler{Handler: handler}).With("service", service)
}

type traceHandler struct{ slog.Handler }

func (h traceHandler) Handle(ctx context.Context, record slog.Record) error {
	if metadata, ok := httpContext.RequestMetadataFromContext(ctx); ok {
		record.AddAttrs(
			slog.String("request_id", metadata.RequestID),
			slog.String("correlation_id", metadata.CorrelationID),
		)
	}
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		record.AddAttrs(
			slog.String("trace_id", spanContext.TraceID().String()),
			slog.String("span_id", spanContext.SpanID().String()),
			slog.Bool("trace_sampled", spanContext.IsSampled()),
		)
	}
	return h.Handler.Handle(ctx, record)
}

func (h traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h traceHandler) WithGroup(name string) slog.Handler {
	return traceHandler{Handler: h.Handler.WithGroup(name)}
}
