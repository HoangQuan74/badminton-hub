package handler

import (
	"net/http"

	authRequest "github.com/kainguyen/badminton-hub/be/gateways/admin/transport/http/request"
	httpMiddleware "github.com/kainguyen/badminton-hub/be/platform/http/middleware"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/gateways/admin/transport/exception"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func RegisterAuth(engine *gin.Engine, client identityV1.IdentityAdminServiceClient) {
	// Build the handler with the owning Identity client.
	handler := NewHandler(client)

	// Register public authentication routes.
	public := engine.Group("/api/v1/admin/auth")
	public.POST("/login", handler.Login)
	public.POST("/refresh", handler.Refresh)
	public.POST("/logout", handler.Logout)

	// Register authenticated administrator routes.
	protected := engine.Group("/api/v1/admin")
	protected.Use(httpMiddleware.RequireBearer)
	protected.GET("/me", handler.CurrentStaff)
	protected.PUT("/me/password", handler.ChangePassword)
}

type Handler struct {
	client identityV1.IdentityAdminServiceClient
}

func NewHandler(client identityV1.IdentityAdminServiceClient) *Handler {
	return &Handler{client: client}
}

func (h *Handler) Login(c *gin.Context) {
	var input authRequest.LoginRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.Login(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.AdminLoginRequest{
			Email:    input.Email,
			Password: input.Password,
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}

func (h *Handler) Refresh(c *gin.Context) {
	var input authRequest.RefreshRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.Refresh(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.AdminRefreshRequest{RefreshToken: input.RefreshToken},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}

func (h *Handler) Logout(c *gin.Context) {
	var input authRequest.RefreshRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.Logout(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.AdminLogoutRequest{RefreshToken: input.RefreshToken},
	)
	exception.WriteGRPC(c, result, err, http.StatusNoContent)
}

func (h *Handler) CurrentStaff(c *gin.Context) {
	result, err := h.client.GetCurrentStaff(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.CurrentStaffRequest{AccessToken: request.BearerToken(c)},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	var input authRequest.ChangePasswordRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.ChangePassword(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.ChangePasswordRequest{
			AccessToken: request.BearerToken(c), CurrentPassword: input.CurrentPassword,
			NewPassword: input.NewPassword,
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusNoContent)
}
