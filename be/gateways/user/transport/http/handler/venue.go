package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/gateways/user/transport/exception"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

var venueErrors = exception.ErrorConfig{NotFoundCode: "VENUE_NOT_FOUND", AlreadyExistsCode: "VENUE_SLUG_ALREADY_EXISTS"}
var courtErrors = exception.ErrorConfig{NotFoundCode: "COURT_NOT_FOUND", AlreadyExistsCode: "COURT_NAME_ALREADY_EXISTS"}

func (h *CatalogHandler) GetPublicVenue(c *gin.Context) {
	result, err := h.catalog.GetPublicVenue(request.RPCContext(c, platformService.UserGateway), &catalogV1.GetPublicVenueRequest{Slug: c.Param("venue_id")})
	exception.WriteGRPC(c, result, err, http.StatusOK, venueErrors)
}

func (h *CatalogHandler) GetTenantVenue(c *gin.Context) {
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(c, nil, err, http.StatusOK, venueErrors)
		return
	}
	result, err := h.catalog.GetTenantVenue(catalogTenantContext(c), &catalogV1.GetTenantVenueRequest{VenueId: c.Param("venue_id")})
	exception.WriteGRPC(c, result, err, http.StatusOK, venueErrors)
}

func (h *CatalogHandler) CreateVenue(c *gin.Context) {
	var input createVenueRequest
	if request.BindJSON(c, &input) {
		return
	}
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(c, nil, err, http.StatusCreated, venueErrors)
		return
	}
	if h.directory == nil {
		exception.WriteGRPC(c, nil, status.Error(codes.Unavailable, "identity unavailable"), http.StatusOK, venueErrors)
		return
	}
	result, err := h.directory.CreateVenue(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantCreateVenueRequest{AccessToken: request.BearerToken(c), OrganizationId: c.Param("organization_id"), Name: input.Name, Slug: input.Slug, Address: input.Address, Timezone: input.Timezone})
	exception.WriteGRPC(c, result, err, http.StatusCreated, venueErrors)
}

func (h *CatalogHandler) UpdateVenue(c *gin.Context) {
	var input updateVenueRequest
	if request.BindJSON(c, &input) {
		return
	}
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(c, nil, err, http.StatusOK, venueErrors)
		return
	}
	if h.directory == nil {
		exception.WriteGRPC(c, nil, status.Error(codes.Unavailable, "identity unavailable"), http.StatusOK, venueErrors)
		return
	}
	result, err := h.directory.UpdateVenue(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantUpdateVenueRequest{AccessToken: request.BearerToken(c), VenueId: c.Param("venue_id"), Name: input.Name, Slug: input.Slug, Address: input.Address, Timezone: input.Timezone})
	exception.WriteGRPC(c, result, err, http.StatusOK, venueErrors)
}

func (h *CatalogHandler) ArchiveVenue(c *gin.Context) {
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(c, nil, err, http.StatusNoContent, venueErrors)
		return
	}
	if h.directory == nil {
		exception.WriteGRPC(c, nil, status.Error(codes.Unavailable, "identity unavailable"), http.StatusOK, venueErrors)
		return
	}
	_, err := h.directory.ArchiveVenue(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantArchiveVenueRequest{AccessToken: request.BearerToken(c), VenueId: c.Param("venue_id")})
	exception.WriteGRPC(c, nil, err, http.StatusNoContent, venueErrors)
}

func (h *CatalogHandler) GetTenantCourt(c *gin.Context) {
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(c, nil, err, http.StatusOK, courtErrors)
		return
	}
	result, err := h.catalog.GetTenantCourt(catalogTenantContext(c), &catalogV1.GetTenantCourtRequest{CourtId: c.Param("court_id")})
	exception.WriteGRPC(c, result, err, http.StatusOK, courtErrors)
}
