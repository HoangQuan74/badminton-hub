package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/gateways/user/transport/exception"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type organizationStaffRequest struct {
	Role   string `json:"role" binding:"required"`
	Status string `json:"status" binding:"required"`
}
type venueStaffRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *CatalogHandler) SetOrganizationStaff(c *gin.Context) {
	var input organizationStaffRequest
	if request.BindJSON(c, &input) {
		return
	}
	if h.directory == nil {
		exception.WriteGRPC(c, nil, status.Error(codes.Unavailable, "identity unavailable"), http.StatusOK, organizationErrors)
		return
	}
	result, err := h.directory.SetOrganizationStaff(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantSetOrganizationStaffRequest{
		AccessToken: request.BearerToken(c), OrganizationId: c.Param("organization_id"), UserId: c.Param("user_id"), Role: input.Role, Status: input.Status,
	})
	exception.WriteGRPC(c, result, err, http.StatusOK, organizationErrors)
}

func (h *CatalogHandler) SetVenueStaff(c *gin.Context) {
	var input venueStaffRequest
	if request.BindJSON(c, &input) {
		return
	}
	if h.directory == nil {
		exception.WriteGRPC(c, nil, status.Error(codes.Unavailable, "identity unavailable"), http.StatusOK, venueErrors)
		return
	}
	result, err := h.directory.SetVenueStaff(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantSetVenueStaffRequest{
		AccessToken: request.BearerToken(c), VenueId: c.Param("venue_id"), UserId: c.Param("user_id"), Status: input.Status,
	})
	exception.WriteGRPC(c, result, err, http.StatusOK, venueErrors)
}
