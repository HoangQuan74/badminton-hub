package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	exception "github.com/kainguyen/badminton-hub/be/gateways/admin/transport/exception"
	roleRequest "github.com/kainguyen/badminton-hub/be/gateways/admin/transport/http/request"
	httpMiddleware "github.com/kainguyen/badminton-hub/be/platform/http/middleware"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func RegisterRole(engine *gin.Engine, client identityV1.IdentityAdminServiceClient) {
	// Build the handler with the owning Identity client.
	handler := NewRoleHandler(client)

	// Register authenticated role routes.
	group := engine.Group("/api/v1/admin/roles")
	group.Use(httpMiddleware.RequireBearer)
	group.GET("/", handler.List)
	group.POST("/", handler.Create)
	group.GET("/:role_id", handler.Get)
	group.PUT("/:role_id", handler.Update)
	group.DELETE("/:role_id", handler.Delete)
	group.PUT("/:role_id/permissions", handler.ReplacePermissions)
}

type RoleHandler struct {
	client identityV1.IdentityAdminServiceClient
}

func NewRoleHandler(client identityV1.IdentityAdminServiceClient) *RoleHandler {
	return &RoleHandler{client: client}
}

func (h *RoleHandler) List(c *gin.Context) {
	result, err := h.client.ListRoles(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.ListRolesRequest{AccessToken: request.BearerToken(c)},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var input roleRequest.RoleRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.CreateRole(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.CreateRoleRequest{
			AccessToken: request.BearerToken(c),
			Name:        input.Name,
			Description: input.Description,
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusCreated)
}

func (h *RoleHandler) Get(c *gin.Context) {
	result, err := h.client.GetRole(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.GetRoleRequest{
			AccessToken: request.BearerToken(c),
			RoleId:      request.Param(c, "role_id"),
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}

func (h *RoleHandler) Update(c *gin.Context) {
	var input roleRequest.RoleRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.UpdateRole(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.UpdateRoleRequest{
			AccessToken: request.BearerToken(c),
			RoleId:      request.Param(c, "role_id"),
			Name:        input.Name,
			Description: input.Description,
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	result, err := h.client.DeleteRole(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.DeleteRoleRequest{
			AccessToken: request.BearerToken(c),
			RoleId:      request.Param(c, "role_id"),
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusNoContent)
}

func (h *RoleHandler) ReplacePermissions(c *gin.Context) {
	var input roleRequest.ReplaceRolePermissionsRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.ReplaceRolePermissions(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.ReplaceRolePermissionsRequest{
			AccessToken:   request.BearerToken(c),
			RoleId:        request.Param(c, "role_id"),
			PermissionIds: input.PermissionIDs,
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}
