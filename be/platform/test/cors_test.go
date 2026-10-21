package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	httpMiddleware "github.com/kainguyen/badminton-hub/be/platform/http/middleware"
)

func TestCORSAllowsConfiguredOriginAndPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(httpMiddleware.CORS("http://localhost:5173"))
	engine.POST("/api", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodOptions, "/api", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "authorization, content-type")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("allow origin = %q", got)
	}
}

func TestCORSRejectsUnknownPreflightOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(httpMiddleware.CORS("http://localhost:5173"))
	engine.POST("/api", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodOptions, "/api", nil)
	request.Header.Set("Origin", "http://evil.example")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected allow origin = %q", got)
	}
}
