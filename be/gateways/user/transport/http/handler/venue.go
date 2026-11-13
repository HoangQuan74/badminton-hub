package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/gateways/user/transport/exception"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type createVenueRequest struct {
	Name     string `json:"name" binding:"required"`
	Slug     string `json:"slug" binding:"required"`
	Address  string `json:"address"`
	Timezone string `json:"timezone"`
}

type updateVenueRequest struct {
	Name     *string `json:"name"`
	Slug     *string `json:"slug"`
	Address  *string `json:"address"`
	Timezone *string `json:"timezone"`
}

func (h *CatalogHandler) GetPublicVenue(c *gin.Context) {
	result, err := h.catalog.GetPublicVenue(request.RPCContext(c, platformService.UserGateway), &catalogV1.GetPublicVenueRequest{Slug: c.Param("venue_id")})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.VenueErrors,
	)
}

func (h *CatalogHandler) GetTenantVenue(c *gin.Context) {
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
			exception.VenueErrors,
		)
		return
	}

	result, err := h.catalog.GetTenantVenue(catalogTenantContext(c), &catalogV1.GetTenantVenueRequest{VenueId: c.Param("venue_id")})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.VenueErrors,
	)
}

func (h *CatalogHandler) GetTenantCourt(c *gin.Context) {
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
			exception.CourtErrors,
		)
		return
	}

	result, err := h.catalog.GetTenantCourt(catalogTenantContext(c), &catalogV1.GetTenantCourtRequest{CourtId: c.Param("court_id")})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.CourtErrors,
	)
}

type VenueHandler struct {
	client identityV1.IdentityVenueServiceClient
}

func RegisterVenue(engine *gin.Engine, client identityV1.IdentityVenueServiceClient) {
	handler := &VenueHandler{client: client}
	group := engine.Group("/api/v1/tenant")
	group.Use(requireTenantBearer)
	group.POST("/:organization_id/venues", handler.CreateVenue)
	group.PATCH("/venues/:venue_id", handler.UpdateVenue)
	group.DELETE("/venues/:venue_id", handler.ArchiveVenue)
}

func (h *VenueHandler) CreateVenue(c *gin.Context) {
	var input createVenueRequest
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

	result, err := h.client.CreateVenue(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantCreateVenueRequest{
		AccessToken:    request.BearerToken(c),
		OrganizationId: c.Param("organization_id"),
		Name:           input.Name,
		Slug:           input.Slug,
		Address:        input.Address,
		Timezone:       input.Timezone,
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusCreated,
		exception.VenueErrors,
	)
}

func (h *VenueHandler) UpdateVenue(c *gin.Context) {
	var input updateVenueRequest
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

	result, err := h.client.UpdateVenue(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantUpdateVenueRequest{
		AccessToken: request.BearerToken(c),
		VenueId:     c.Param("venue_id"),
		Name:        input.Name,
		Slug:        input.Slug,
		Address:     input.Address,
		Timezone:    input.Timezone,
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.VenueErrors,
	)
}

func (h *VenueHandler) ArchiveVenue(c *gin.Context) {
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

	_, err := h.client.ArchiveVenue(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantArchiveVenueRequest{
		AccessToken: request.BearerToken(c),
		VenueId:     c.Param("venue_id"),
	})
	exception.WriteGRPC(
		c,
		nil,
		err,
		http.StatusNoContent,
		exception.VenueErrors,
	)
}
