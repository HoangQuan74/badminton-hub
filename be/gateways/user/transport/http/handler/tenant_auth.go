package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/gateways/user/transport/exception"
	authRequest "github.com/kainguyen/badminton-hub/be/gateways/user/transport/http/request"
	httpMiddleware "github.com/kainguyen/badminton-hub/be/platform/http/middleware"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func RegisterTenantAuth(engine *gin.Engine, client identityV1.IdentityTenantAuthServiceClient) {
	handler := &TenantAuthHandler{client: client}
	public := engine.Group("/api/v1/tenant/auth")
	public.POST("/login", handler.Login)
	public.POST("/refresh", handler.Refresh)
	public.POST("/logout", handler.Logout)
	protected := engine.Group("/api/v1/tenant")
	protected.Use(httpMiddleware.RequireBearer)
	protected.GET("/me", handler.CurrentUser)
}

type TenantAuthHandler struct {
	client identityV1.IdentityTenantAuthServiceClient
}

func (h *TenantAuthHandler) Login(c *gin.Context) {
	var input authRequest.CredentialsRequest
	if request.BindJSON(c, &input) {
		return
	}

	result, err := h.client.Login(
		request.RPCContext(c, platformService.UserGateway),
		&identityV1.LoginRequest{
			Email:    input.Email,
			Password: input.Password,
		},
	)
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
	)
}

func (h *TenantAuthHandler) Refresh(c *gin.Context) {
	var input authRequest.RefreshRequest
	if request.BindJSON(c, &input) {
		return
	}

	result, err := h.client.Refresh(
		request.RPCContext(c, platformService.UserGateway),
		&identityV1.RefreshRequest{RefreshToken: input.RefreshToken},
	)
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
	)
}

func (h *TenantAuthHandler) Logout(c *gin.Context) {
	var input authRequest.RefreshRequest
	if request.BindJSON(c, &input) {
		return
	}

	_, err := h.client.Logout(
		request.RPCContext(c, platformService.UserGateway),
		&identityV1.LogoutRequest{RefreshToken: input.RefreshToken},
	)
	exception.WriteGRPC(
		c,
		nil,
		err,
		http.StatusNoContent,
	)
}

func (h *TenantAuthHandler) CurrentUser(c *gin.Context) {
	token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	result, err := h.client.GetCurrentUser(
		request.RPCContext(c, platformService.UserGateway),
		&identityV1.GetCurrentUserRequest{AccessToken: token},
	)
	exception.WriteGRPC(
		c,
		meResponse{User: result},
		err,
		http.StatusOK,
	)
}
