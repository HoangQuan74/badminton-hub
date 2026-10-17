package test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	platformMetrics "github.com/kainguyen/badminton-hub/be/platform/metrics"
)

func TestMiddlewareAndHandlerExposeBoundedHTTPMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registry := platformMetrics.New("test-gateway")
	engine := gin.New()
	engine.Use(registry.Middleware())
	engine.GET("/users/:id", func(c *gin.Context) { c.Status(201) })

	request := httptest.NewRequest("GET", "/users/42", nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != 201 {
		t.Fatalf("status = %d, want 201", response.Code)
	}

	metricsResponse := httptest.NewRecorder()
	registry.Handler().ServeHTTP(metricsResponse, httptest.NewRequest("GET", "/metrics", nil))
	body := metricsResponse.Body.String()
	for _, expected := range []string{
		`http_requests_total{service="test-gateway",method="GET",route="/users/:id",status="201"} 1`,
		`http_request_duration_seconds_count{service="test-gateway",method="GET",route="/users/:id"} 1`,
		`http_in_flight_requests{service="test-gateway"} 0`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("metrics missing %q in %s", expected, body)
		}
	}
}
