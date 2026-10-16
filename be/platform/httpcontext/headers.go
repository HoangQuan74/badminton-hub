package httpcontext

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestContextHeaders propagates bounded request and correlation identifiers
// across public and internal service boundaries.
func RequestContextHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" || len(requestID) > 128 {
			requestID = uuid.NewString()
		}
		correlationID := strings.TrimSpace(c.GetHeader("X-Correlation-ID"))
		if correlationID == "" || len(correlationID) > 128 {
			correlationID = requestID
		}
		c.Header("X-Request-ID", requestID)
		c.Header("X-Correlation-ID", correlationID)
		c.Request.Header.Set("X-Request-ID", requestID)
		c.Request.Header.Set("X-Correlation-ID", correlationID)
		c.Next()
	}
}
