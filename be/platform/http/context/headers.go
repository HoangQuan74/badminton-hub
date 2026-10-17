package httpContext

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type metadataKey struct{}

type RequestMetadata struct {
	RequestID     string
	CorrelationID string
}

func WithRequestMetadata(ctx context.Context, metadata RequestMetadata) context.Context {
	return context.WithValue(ctx, metadataKey{}, metadata)
}

func RequestMetadataFromContext(ctx context.Context) (RequestMetadata, bool) {
	metadata, ok := ctx.Value(metadataKey{}).(RequestMetadata)
	return metadata, ok
}

// RequestContextHeaders creates request metadata at a public boundary. Client
// supplied identifiers are deliberately ignored so callers cannot spoof the
// identifiers used for tracing and operational logs.
func RequestContextHeaders() gin.HandlerFunc {
	return requestContextHeaders(false)
}

// TrustedRequestContextHeaders propagates identifiers created by a trusted
// upstream, such as the public Gateway forwarding to a private service.
func TrustedRequestContextHeaders() gin.HandlerFunc {
	return requestContextHeaders(true)
}

func requestContextHeaders(acceptIncoming bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := ""
		if acceptIncoming {
			requestID = validHeaderID(c.GetHeader("X-Request-ID"))
		}
		if requestID == "" {
			requestID = uuid.NewString()
		}
		correlationID := ""
		if acceptIncoming {
			correlationID = validHeaderID(c.GetHeader("X-Correlation-ID"))
		}
		if correlationID == "" {
			correlationID = requestID
		}
		c.Header("X-Request-ID", requestID)
		c.Header("X-Correlation-ID", correlationID)
		c.Request.Header.Set("X-Request-ID", requestID)
		c.Request.Header.Set("X-Correlation-ID", correlationID)
		c.Request = c.Request.WithContext(WithRequestMetadata(c.Request.Context(), RequestMetadata{
			RequestID: requestID, CorrelationID: correlationID,
		}))
		c.Next()
	}
}

func validHeaderID(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > 128 {
		return ""
	}
	for i := 0; i < len(value); i++ {
		character := value[i]
		if !((character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '.' || character == '_' || character == ':' || character == '-') {
			return ""
		}
	}
	return value
}
