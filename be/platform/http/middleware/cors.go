package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	allowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	allowHeaders = "Accept, Accept-Language, Authorization, Content-Type, X-Request-ID, X-Correlation-ID"
)

// CORS allows browser clients from the configured, comma-separated origins.
// An empty Origin is allowed through because same-origin and non-browser
// clients do not send CORS headers.
func CORS(configuredOrigins string) gin.HandlerFunc {
	allowedOrigins := make(map[string]struct{})
	for _, rawOrigin := range strings.Split(configuredOrigins, ",") {
		origin := strings.TrimSpace(rawOrigin)
		if origin != "" {
			allowedOrigins[origin] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin == "" {
			c.Next()
			return
		}

		if _, ok := allowedOrigins[origin]; !ok {
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", allowMethods)
		c.Header("Access-Control-Allow-Headers", allowHeaders)
		c.Header("Access-Control-Expose-Headers", "X-Request-ID, X-Correlation-ID")
		c.Header("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
