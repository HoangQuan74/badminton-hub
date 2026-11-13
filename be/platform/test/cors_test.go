package test

import (
	"net/http"
	"net/http/httptest"
	"strings"
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
	if !strings.Contains(response.Header().Get("Access-Control-Allow-Headers"), "Idempotency-Key") {
		t.Fatal("browser cannot send idempotency key")
	}
	for _, header := range []string{"Location", "Idempotency-Replayed", "Idempotency-Key-Expires-At"} {
		if !strings.Contains(response.Header().Get("Access-Control-Expose-Headers"), header) {
			t.Fatalf("browser cannot read %s", header)
		}
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

func TestCORSActualRequestsAndReplayHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, origin := range []string{"", "http://localhost:5173", "http://other.example"} {
		engine := gin.New()
		engine.Use(httpMiddleware.CORS(" , http://localhost:5173, "))
		engine.POST("/api", func(c *gin.Context) { c.Status(http.StatusCreated) })
		request := httptest.NewRequest(http.MethodPost, "/api", nil)
		request.Header.Set("Origin", origin)
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, request)
		if response.Code != http.StatusCreated {
			t.Fatalf("actual request was blocked: %s", origin)
		}
		allowed := response.Header().Get("Access-Control-Allow-Origin")
		if origin == "http://localhost:5173" && allowed != origin {
			t.Fatal("configured origin not allowed")
		}
		if origin != "http://localhost:5173" && allowed != "" {
			t.Fatal("unexpected cross-origin access")
		}
	}
}
