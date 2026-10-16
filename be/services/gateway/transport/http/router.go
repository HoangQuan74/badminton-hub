package gatewayhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/platform/httpcontext"
	gatewayconfig "github.com/kainguyen/badminton-hub/be/services/gateway/config"
)

type Dependencies struct {
	Config     gatewayconfig.Config
	HTTPClient *http.Client
	Transport  http.RoundTripper
}

// New creates the public Gateway/BFF. Business routes are forwarded to the
// owning service; the Gateway does not own business storage or state.
func New(dependencies Dependencies) (*gin.Engine, error) {
	engine := newEngine()
	transport := dependencies.Transport
	if transport == nil {
		transport = gatewayTransport(dependencies.Config.DependencyTimeout)
	}
	proxy, err := identityProxy(dependencies.Config.IdentityURL, transport)
	if err != nil {
		return nil, err
	}
	engine.Any("/api/v1/auth/*path", proxy)
	engine.Any("/api/v1/me", proxy)
	registerInfrastructureRoutes(engine, dependencies)
	return engine, nil
}

func newEngine() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), httpcontext.RequestContextHeaders())
	return engine
}

func identityProxy(rawURL string, transport http.RoundTripper) (gin.HandlerFunc, error) {
	target, err := url.Parse(rawURL)
	if err != nil || target.Scheme == "" || target.Host == "" {
		if err == nil {
			err = fmt.Errorf("URL must include scheme and host")
		}
		return nil, fmt.Errorf("parse GATEWAY_IDENTITY_URL: %w", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport
	proxy.ErrorHandler = func(writer http.ResponseWriter, _ *http.Request, _ error) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte(`{"error":{"code":"IDENTITY_UNAVAILABLE","message":"Identity service is unavailable."}}`))
	}
	return gin.WrapH(proxy), nil
}

func gatewayTransport(timeout time.Duration) *http.Transport {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: time.Second,
	}
}

type statusResponse struct {
	Status string `json:"status"`
}

func registerInfrastructureRoutes(engine *gin.Engine, dependencies Dependencies) {
	client := dependencies.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: dependencies.Config.DependencyTimeout}
	}
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, statusResponse{Status: "ok"})
	})
	engine.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, dependencies.Config.IdentityURL+"/ready", nil)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, statusResponse{Status: "not_ready"})
			return
		}
		result, err := client.Do(request)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, statusResponse{Status: "not_ready"})
			return
		}
		defer result.Body.Close()
		if result.StatusCode < http.StatusOK || result.StatusCode >= http.StatusMultipleChoices {
			c.JSON(http.StatusServiceUnavailable, statusResponse{Status: "not_ready"})
			return
		}
		c.JSON(http.StatusOK, statusResponse{Status: "ready"})
	})
}
