package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/gateways/user/transport/exception"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type StaffHandler struct {
	client identityV1.IdentityStaffServiceClient
}

func RegisterStaff(engine *gin.Engine, client identityV1.IdentityStaffServiceClient) {
	handler := &StaffHandler{client: client}
	group := engine.Group("/api/v1/tenant")
	group.Use(requireTenantBearer)
	group.GET("/:organization_id/staff", handler.ListOrganizationStaff)
	group.POST("/:organization_id/staff", handler.AddOrganizationStaff)
	group.DELETE("/:organization_id/staff/:user_id", handler.RemoveOrganizationStaff)
	group.PUT("/:organization_id/staff/:user_id", handler.SetOrganizationStaff)
	group.PUT("/venues/:venue_id/staff/:user_id", handler.SetVenueStaff)
}

type organizationStaffRequest struct {
	Role   string `json:"role" binding:"required"`
	Status string `json:"status" binding:"required"`
}
type venueStaffRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *StaffHandler) SetOrganizationStaff(c *gin.Context) {
	var input organizationStaffRequest
	if request.BindJSON(c, &input) {
		return
	}

	if h.client == nil {
		exception.WriteGRPC(
			c,
			nil,
			exception.ErrGRPCIdentityUnavailable,
			http.StatusOK,
			exception.OrganizationErrors,
		)
		return
	}

	result, err := h.client.SetOrganizationStaff(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantSetOrganizationStaffRequest{
		AccessToken:    request.BearerToken(c),
		OrganizationId: c.Param("organization_id"),
		UserId:         c.Param("user_id"),
		Role:           input.Role,
		Status:         input.Status,
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.OrganizationErrors,
	)
}

func (h *StaffHandler) SetVenueStaff(c *gin.Context) {
	var input venueStaffRequest
	if request.BindJSON(c, &input) {
		return
	}

	if h.client == nil {
		exception.WriteGRPC(
			c,
			nil,
			exception.ErrGRPCIdentityUnavailable,
			http.StatusOK,
			exception.VenueErrors,
		)
		return
	}

	result, err := h.client.SetVenueStaff(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantSetVenueStaffRequest{
		AccessToken: request.BearerToken(c),
		VenueId:     c.Param("venue_id"),
		UserId:      c.Param("user_id"),
		Status:      input.Status,
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.VenueErrors,
	)
}

func (h *StaffHandler) ListOrganizationStaff(c *gin.Context) {
	result, err := h.client.ListOrganizationStaff(request.RPCContext(c, platformService.UserGateway), &identityV1.ListOrganizationStaffRequest{
		AccessToken:    request.BearerToken(c),
		OrganizationId: c.Param("organization_id"),
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.OrganizationErrors,
	)
}

func (h *StaffHandler) AddOrganizationStaff(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role" binding:"required,oneof=tenant_staff venue_staff"`
	}
	if request.BindJSON(c, &input) {
		return
	}

	result, err := h.client.AddOrganizationStaff(request.RPCContext(c, platformService.UserGateway), &identityV1.AddOrganizationStaffRequest{
		AccessToken:    request.BearerToken(c),
		OrganizationId: c.Param("organization_id"),
		Email:          input.Email,
		Role:           input.Role,
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusCreated,
		exception.OrganizationErrors,
	)
}

func (h *StaffHandler) RemoveOrganizationStaff(c *gin.Context) {
	_, err := h.client.RemoveOrganizationStaff(request.RPCContext(c, platformService.UserGateway), &identityV1.RemoveOrganizationStaffRequest{
		AccessToken:    request.BearerToken(c),
		OrganizationId: c.Param("organization_id"),
		UserId:         c.Param("user_id"),
	})
	exception.WriteGRPC(
		c,
		nil,
		err,
		http.StatusNoContent,
		exception.OrganizationErrors,
	)
}
