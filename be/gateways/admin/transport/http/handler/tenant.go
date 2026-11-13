package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/gateways/admin/transport/exception"
	httpMiddleware "github.com/kainguyen/badminton-hub/be/platform/http/middleware"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type TenantHandler struct {
	client identityV1.IdentityTenantServiceClient
}

func RegisterTenant(engine *gin.Engine, client identityV1.IdentityTenantServiceClient) {
	handler := &TenantHandler{client: client}
	group := engine.Group("/api/v1/admin/tenants")
	group.Use(httpMiddleware.RequireBearer)
	group.GET("", handler.ListRegistrations)
	group.POST("/:organization_id/approve", handler.Approve)
	group.POST("/:organization_id/reject", handler.Reject)
}

var tenantErrors = exception.ErrorConfig{
	NotFoundCode:           "TENANT_NOT_FOUND",
	FailedPreconditionCode: "TENANT_STATE_CONFLICT",
}

func (h *TenantHandler) ListRegistrations(c *gin.Context) {
	result, err := h.client.ListRegistrations(request.RPCContext(c, platformService.AdminGateway), &identityV1.TenantListRegistrationsRequest{
		AccessToken: request.BearerToken(c),
		Status:      c.Query("status"),
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		tenantErrors,
	)
}

func (h *TenantHandler) Approve(c *gin.Context) {
	h.review(c, "ACTIVE")
}

func (h *TenantHandler) Reject(c *gin.Context) {
	h.review(c, "REJECTED")
}

func (h *TenantHandler) review(c *gin.Context, decision string) {
	var input struct {
		Note string `json:"note" binding:"max=1000"`
	}
	if request.BindJSON(c, &input) {
		return
	}

	result, err := h.client.ReviewRegistration(request.RPCContext(c, platformService.AdminGateway), &identityV1.TenantReviewRegistrationRequest{
		AccessToken:    request.BearerToken(c),
		OrganizationId: c.Param("organization_id"),
		Decision:       decision,
		Note:           input.Note,
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		tenantErrors,
	)
}
