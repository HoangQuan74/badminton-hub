package test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	httpContext "github.com/kainguyen/badminton-hub/be/platform/http/context"
)

func TestRequestContextHeadersGeneratesServerOwnedIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(httpContext.RequestContextHeaders())
	engine.GET("/", func(c *gin.Context) {
		if got := c.GetHeader("X-Request-ID"); got == "client-request" {
			t.Fatal("public middleware accepted a client-supplied request ID")
		}
		if got := c.GetHeader("X-Correlation-ID"); got == "client-correlation" {
			t.Fatal("public middleware accepted a client-supplied correlation ID")
		}
		c.Status(204)
	})

	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("X-Request-ID", "client-request")
	request.Header.Set("X-Correlation-ID", "client-correlation")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != 204 {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if response.Header().Get("X-Request-ID") == "" || response.Header().Get("X-Correlation-ID") == "" {
		t.Fatal("server did not emit request metadata")
	}
}

func TestTrustedRequestContextHeadersPropagatesValidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(httpContext.TrustedRequestContextHeaders())
	engine.GET("/", func(c *gin.Context) {
		if got := c.GetHeader("X-Request-ID"); got != "trusted-request" {
			t.Fatalf("request ID = %q, want trusted-request", got)
		}
		if got := c.GetHeader("X-Correlation-ID"); got != "trusted-correlation" {
			t.Fatalf("correlation ID = %q, want trusted-correlation", got)
		}
		c.Status(204)
	})

	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("X-Request-ID", "trusted-request")
	request.Header.Set("X-Correlation-ID", "trusted-correlation")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != 204 {
		t.Fatalf("status = %d, want 204", response.Code)
	}
}

func TestTrustedRequestContextHeadersRejectsInvalidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(httpContext.TrustedRequestContextHeaders())
	engine.GET("/", func(c *gin.Context) { c.Status(204) })

	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("X-Request-ID", "bad id")
	request.Header.Set("X-Correlation-ID", "bad\tid")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != 204 {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if response.Header().Get("X-Request-ID") == "bad id" || response.Header().Get("X-Correlation-ID") == "bad\tid" {
		t.Fatal("trusted middleware propagated an invalid ID")
	}
}

func TestTrustedRequestContextHeadersRejectsOversizedIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(httpContext.TrustedRequestContextHeaders())
	engine.GET("/", func(c *gin.Context) { c.Status(204) })
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("X-Request-ID", strings.Repeat("a", 129))
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != 204 || response.Header().Get("X-Request-ID") == strings.Repeat("a", 129) {
		t.Fatal("trusted middleware accepted an oversized request ID")
	}
}
