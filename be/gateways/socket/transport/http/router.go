package http

import (
	stdcontext "context"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"

	"github.com/gin-gonic/gin"
	socketConfig "github.com/kainguyen/badminton-hub/be/gateways/socket/config"
	socketDocs "github.com/kainguyen/badminton-hub/be/gateways/socket/docs"
	handler "github.com/kainguyen/badminton-hub/be/gateways/socket/transport/http/handler"
	grpcClient "github.com/kainguyen/badminton-hub/be/platform/grpc/client"
	httpContext "github.com/kainguyen/badminton-hub/be/platform/http/context"
	"github.com/kainguyen/badminton-hub/be/platform/metrics"
	socketV1 "github.com/kainguyen/badminton-hub/contracts/gen/socket/v1"
)

type Dependencies struct{ Config socketConfig.Config }

func New(dependencies Dependencies) (*gin.Engine, error) {
	connection, err := grpcClient.Dial(stdcontext.Background(), dependencies.Config.GRPCTarget, grpcClient.Options{DialTimeout: dependencies.Config.DependencyTimeout, ServiceIdentity: platformService.SocketGateway, Block: true})
	if err != nil {
		return nil, err
	}
	engine := gin.New()
	registry := metrics.New(platformService.SocketGateway)
	engine.Use(gin.Logger(), gin.Recovery(), httpContext.RequestContextHeaders(), registry.Middleware())
	handler.RegisterWebSocket(engine, socketV1.NewSocketServiceClient(connection), registry.Handler())
	socketDocs.Register(engine)
	return engine, nil
}
