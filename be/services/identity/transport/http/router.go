package identityhttp

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/platform/httpcontext"
	identityconfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	identityhandler "github.com/kainguyen/badminton-hub/be/services/identity/transport/http/handler"
	localization "github.com/kainguyen/badminton-hub/be/services/identity/transport/http/i18n"
	"github.com/kainguyen/badminton-hub/be/services/identity/transport/http/models/response"
)

type Dependencies struct {
	Config identityconfig.Config
	Pool   *pgxpool.Pool
	Logger *slog.Logger
}

// New creates the Identity service HTTP surface. Identity owns credentials,
// refresh sessions, token issuance, and the /me query.
func New(dependencies Dependencies) (*gin.Engine, error) {
	translator, err := localization.New()
	if err != nil {
		return nil, err
	}
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), httpcontext.RequestContextHeaders(), translator.Middleware())
	registerInfrastructureRoutes(engine, dependencies)
	identityhandler.RegisterAuthRoutes(engine, identityhandler.AuthDependencies{Config: dependencies.Config, Pool: dependencies.Pool, Logger: dependencies.Logger})
	return engine, nil
}

func registerInfrastructureRoutes(engine *gin.Engine, dependencies Dependencies) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.StatusResponse{Status: "ok"})
	})
	engine.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := dependencies.Pool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, response.StatusResponse{Status: "not_ready"})
			return
		}
		c.JSON(http.StatusOK, response.StatusResponse{Status: "ready"})
	})
}
