package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	exception "github.com/kainguyen/badminton-hub/be/gateways/admin/transport/exception"
	permissionRequest "github.com/kainguyen/badminton-hub/be/gateways/admin/transport/http/request"
	httpMiddleware "github.com/kainguyen/badminton-hub/be/platform/http/middleware"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func RegisterPermission(engine *gin.Engine, client identityV1.IdentityPermissionServiceClient) {
	handler := NewPermissionHandler(client)
	group := engine.Group("/api/v1/admin/permissions")
	group.Use(httpMiddleware.RequireBearer)
	group.GET("/", handler.List)
	group.POST("/", handler.Create)
	group.GET("/:permission_id", handler.Get)
	group.PUT("/:permission_id", handler.Update)
	group.DELETE("/:permission_id", handler.Delete)
}

type PermissionHandler struct {
	client identityV1.IdentityPermissionServiceClient
}

func NewPermissionHandler(client identityV1.IdentityPermissionServiceClient) *PermissionHandler {
	return &PermissionHandler{client: client}
}

func (h *PermissionHandler) List(c *gin.Context) {
	// Forward to the owning permission service.
	result, err := h.client.ListPermissions(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.ListPermissionsRequest{AccessToken: request.BearerToken(c)},
	)
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.PermissionErrors,
	)
}

func (h *PermissionHandler) Create(c *gin.Context) {
	// Bind the permission replacement values.
	var input permissionRequest.PermissionRequest
	if request.BindJSON(c, &input) {
		return
	}

	// Forward to the owning permission service.
	result, err := h.client.CreatePermission(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.CreatePermissionRequest{
			AccessToken: request.BearerToken(c),
			Module:      input.Module,
			Path:        input.Path,
			Action:      input.Action,
			Description: input.Description,
		},
	)
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusCreated,
		exception.PermissionErrors,
	)
}

func (h *PermissionHandler) Get(c *gin.Context) {
	// Forward to the owning permission service.
	result, err := h.client.GetPermission(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.GetPermissionRequest{AccessToken: request.BearerToken(c),
			PermissionId: request.Param(c, "permission_id")},
	)
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.PermissionErrors,
	)
}

func (h *PermissionHandler) Update(c *gin.Context) {
	// Bind the permission replacement values.
	var input permissionRequest.PermissionRequest
	if request.BindJSON(c, &input) {
		return
	}

	// Forward to the owning permission service.
	result, err := h.client.UpdatePermission(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.UpdatePermissionRequest{
			AccessToken:  request.BearerToken(c),
			Module:       input.Module,
			Path:         input.Path,
			Action:       input.Action,
			Description:  input.Description,
			PermissionId: request.Param(c, "permission_id"),
		},
	)
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.PermissionErrors,
	)
}

func (h *PermissionHandler) Delete(c *gin.Context) {
	// Forward to the owning permission service.
	result, err := h.client.DeletePermission(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.DeletePermissionRequest{AccessToken: request.BearerToken(c),
			PermissionId: request.Param(c, "permission_id")},
	)
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusNoContent,
		exception.PermissionErrors,
	)
}
