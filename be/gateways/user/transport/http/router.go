package http

import (
	"context"

	platformService "github.com/kainguyen/badminton-hub/be/platform/service"

	"github.com/gin-gonic/gin"
	userConfig "github.com/kainguyen/badminton-hub/be/gateways/user/config"
	userDocs "github.com/kainguyen/badminton-hub/be/gateways/user/docs"
	handler "github.com/kainguyen/badminton-hub/be/gateways/user/transport/http/handler"
	grpcClient "github.com/kainguyen/badminton-hub/be/platform/grpc/client"
	httpContext "github.com/kainguyen/badminton-hub/be/platform/http/context"
	httpMiddleware "github.com/kainguyen/badminton-hub/be/platform/http/middleware"
	"github.com/kainguyen/badminton-hub/be/platform/metrics"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type Dependencies struct {
	Config userConfig.Config
}

func New(dependencies Dependencies) (*gin.Engine, error) {
	connection, err := grpcClient.Dial(context.Background(), dependencies.Config.IdentityGRPCTarget, grpcClient.Options{DialTimeout: dependencies.Config.DependencyTimeout, ServiceIdentity: platformService.UserGateway, Block: true})
	if err != nil {
		return nil, err
	}
	catalogConnection, err := grpcClient.Dial(context.Background(), dependencies.Config.CatalogGRPCTarget, grpcClient.Options{DialTimeout: dependencies.Config.DependencyTimeout, ServiceIdentity: platformService.UserGateway, Block: true})
	if err != nil {
		return nil, err
	}
	engine := gin.New()
	registry := metrics.New(platformService.UserGateway)
	engine.Use(
		gin.Logger(),
		gin.Recovery(),
		httpMiddleware.CORS(dependencies.Config.CORSAllowedOrigins),
		httpContext.RequestContextHeaders(),
		registry.Middleware(),
	)

	identityClient := identityV1.NewIdentityAuthServiceClient(connection)
	handler.RegisterAuth(engine, identityClient)
	handler.RegisterTenant(engine, identityV1.NewIdentityTenantServiceClient(connection), dependencies.Config.DependencyTimeout)
	handler.RegisterVenue(engine, identityV1.NewIdentityVenueServiceClient(connection))
	handler.RegisterStaff(engine, identityV1.NewIdentityStaffServiceClient(connection))
	handler.RegisterTenantAuth(engine, identityV1.NewIdentityTenantAuthServiceClient(connection))
	handler.RegisterCatalog(engine, identityV1.NewIdentityTenantAuthServiceClient(connection), catalogV1.NewCatalogServiceClient(catalogConnection))

	engine.GET("/health", health)
	engine.GET("/ready", ready)
	engine.GET("/metrics", gin.WrapH(registry.Handler()))
	userDocs.Register(engine)

	return engine, nil
}

func health(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) }
func ready(c *gin.Context)  { c.JSON(200, gin.H{"status": "ready"}) }
