package handler

import (
	httpMiddleware "github.com/kainguyen/badminton-hub/be/platform/http/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	exception "github.com/kainguyen/badminton-hub/be/gateways/admin/transport/exception"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func RegisterPermission(engine *gin.Engine, client identityV1.IdentityAdminServiceClient) {
	// Build the handler with the owning Identity client.
	handler := NewPermissionHandler(client)

	// Register the authenticated permission catalog route.
	group := engine.Group("/api/v1/admin/permissions")
	group.Use(httpMiddleware.RequireBearer)
	group.GET("/", handler.List)
}

type PermissionHandler struct {
	client identityV1.IdentityAdminServiceClient
}

func NewPermissionHandler(client identityV1.IdentityAdminServiceClient) *PermissionHandler {
	return &PermissionHandler{client: client}
}

func (h *PermissionHandler) List(c *gin.Context) {
	result, err := h.client.ListPermissions(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.ListPermissionsRequest{
			AccessToken: request.BearerToken(c),
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}
