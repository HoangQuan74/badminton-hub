package http

import (
	"context"

	"github.com/gin-gonic/gin"
	adminConfig "github.com/kainguyen/badminton-hub/be/gateways/admin/config"
	adminDocs "github.com/kainguyen/badminton-hub/be/gateways/admin/docs"
	handler "github.com/kainguyen/badminton-hub/be/gateways/admin/transport/http/handler"
	grpcClient "github.com/kainguyen/badminton-hub/be/platform/grpc/client"
	httpContext "github.com/kainguyen/badminton-hub/be/platform/http/context"
	"github.com/kainguyen/badminton-hub/be/platform/metrics"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type Dependencies struct {
	Config adminConfig.Config
}

func New(dependencies Dependencies) (*gin.Engine, error) {
	connection, err := grpcClient.Dial(context.Background(), dependencies.Config.IdentityGRPCTarget, grpcClient.Options{DialTimeout: dependencies.Config.DependencyTimeout, ServiceIdentity: platformService.AdminGateway, Block: true})
	if err != nil {
		return nil, err
	}
	engine := gin.New()
	registry := metrics.New(platformService.AdminGateway)
	engine.Use(gin.Logger(), gin.Recovery(), httpContext.RequestContextHeaders(), registry.Middleware())
	client := identityV1.NewIdentityAdminServiceClient(connection)

	handler.RegisterAuth(engine, client)
	handler.RegisterStaff(engine, client)
	handler.RegisterRole(engine, client)
	handler.RegisterPermission(engine, client)

	engine.GET("/health", health)
	engine.GET("/ready", ready)
	engine.GET("/metrics", gin.WrapH(registry.Handler()))

	adminDocs.Register(engine)
	return engine, nil
}

func health(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) }
func ready(c *gin.Context)  { c.JSON(200, gin.H{"status": "ready"}) }
