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

func RegisterAuth(engine *gin.Engine, client identityV1.IdentityAuthServiceClient) {
	// Build the handler with the owning Identity client.
	handler := NewHandler(client)

	// Register public authentication routes.
	public := engine.Group("/api/v1/auth")
	public.POST("/register", handler.Register)
	public.POST("/login", handler.Login)
	public.POST("/refresh", handler.Refresh)
	public.POST("/logout", handler.Logout)

	// Register the authenticated current-user route.
	protected := engine.Group("/api/v1")
	protected.Use(httpMiddleware.RequireBearer)
	protected.GET("/me", handler.CurrentUser)
}

type Handler struct {
	client identityV1.IdentityAuthServiceClient
}

func NewHandler(client identityV1.IdentityAuthServiceClient) *Handler {
	return &Handler{client: client}
}
func (h *Handler) Register(c *gin.Context) {
	var input authRequest.RegisterRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.Register(
		request.RPCContext(c, platformService.UserGateway),
		&identityV1.RegisterRequest{
			Email:    input.Email,
			Password: input.Password,
			FullName: input.FullName,
			Phone:    input.Phone,
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusCreated)
}
func (h *Handler) Login(c *gin.Context) {
	var input authRequest.CredentialsRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.Login(
		request.RPCContext(c, platformService.UserGateway),
		&identityV1.LoginRequest{Email: input.Email, Password: input.Password},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}
func (h *Handler) Refresh(c *gin.Context) {
	var input authRequest.RefreshRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.Refresh(
		request.RPCContext(c, platformService.UserGateway),
		&identityV1.RefreshRequest{RefreshToken: input.RefreshToken},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}
func (h *Handler) Logout(c *gin.Context) {
	var input authRequest.RefreshRequest
	if request.BindJSON(c, &input) {
		return
	}
	_, err := h.client.Logout(
		request.RPCContext(c, platformService.UserGateway),
		&identityV1.LogoutRequest{RefreshToken: input.RefreshToken},
	)
	exception.WriteGRPC(c, nil, err, http.StatusNoContent)
}
func (h *Handler) CurrentUser(c *gin.Context) {
	token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	result, err := h.client.GetCurrentUser(
		request.RPCContext(c, platformService.UserGateway),
		&identityV1.GetCurrentUserRequest{AccessToken: token},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}
