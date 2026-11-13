package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/gateways/user/transport/exception"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc/metadata"
)

type CatalogHandler struct {
	identity identityV1.IdentityTenantAuthServiceClient
	catalog  catalogV1.CatalogServiceClient
}

type courtRequest struct {
	Name                string `json:"name" binding:"required"`
	SlotIntervalMinutes int32  `json:"slot_interval_minutes"`
	Status              string `json:"status"`
}

func RegisterCatalog(engine *gin.Engine, identity identityV1.IdentityTenantAuthServiceClient, catalog catalogV1.CatalogServiceClient) {
	handler := &CatalogHandler{
		identity: identity,
		catalog:  catalog,
	}

	engine.GET("/api/v1/venues", handler.ListPublicVenues)
	engine.GET("/api/v1/venues/:venue_id", handler.GetPublicVenue)
	engine.GET("/api/v1/venues/:venue_id/courts", handler.ListPublicVenueCourts)
	engine.GET("/api/v1/venues/:venue_id/images", handler.ListPublicVenueImages)
	engine.GET("/api/v1/venue-images/:image_id", handler.GetPublicVenueImage)

	tenant := engine.Group("/api/v1/tenant")
	tenant.Use(requireTenantBearer)
	tenant.GET("/venues", handler.ListTenantVenues)
	tenant.GET("/venues/:venue_id", handler.GetTenantVenue)
	tenant.GET("/courts/:court_id", handler.GetTenantCourt)
	tenant.GET("/venues/:venue_id/courts", handler.ListTenantCourts)
	tenant.GET("/venues/:venue_id/images", handler.ListTenantVenueImages)
	tenant.POST("/venues/:venue_id/images", handler.UploadVenueImage)
	tenant.GET("/venue-images/:image_id", handler.GetTenantVenueImage)
	tenant.POST("/venues/:venue_id/courts", handler.CreateCourt)
	tenant.PATCH("/courts/:court_id", handler.UpdateCourt)
	tenant.DELETE("/courts/:court_id", handler.ArchiveCourt)
}

func requireTenantBearer(c *gin.Context) {
	if request.BearerToken(c) == "" {
		exception.AbortUnauthorized(c)
		return
	}

	c.Next()
}

func (h *CatalogHandler) ListPublicVenues(c *gin.Context) {
	result, err := h.catalog.ListPublicVenues(request.RPCContext(c, platformService.UserGateway), &catalogV1.ListPublicVenuesRequest{})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.VenueErrors,
	)
}

func (h *CatalogHandler) ListPublicVenueCourts(c *gin.Context) {
	result, err := h.catalog.ListPublicVenueCourts(request.RPCContext(c, platformService.UserGateway), &catalogV1.ListPublicVenueCourtsRequest{VenueId: c.Param("venue_id")})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.VenueErrors,
	)
}

func (h *CatalogHandler) ListTenantVenues(c *gin.Context) {
	actorID, err := h.tenantActorID(c)
	if err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
		)
		return
	}

	result, err := h.catalog.ListTenantVenues(catalogTenantContext(c), &catalogV1.ListTenantVenuesRequest{ActorUserId: actorID})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
	)
}

func (h *CatalogHandler) ListTenantCourts(c *gin.Context) {
	actorID, err := h.tenantActorID(c)
	if err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
		)
		return
	}

	result, err := h.catalog.ListTenantCourts(catalogTenantContext(c), &catalogV1.ListTenantCourtsRequest{
		ActorUserId: actorID,
		VenueId:     c.Param("venue_id"),
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
	)
}

func (h *CatalogHandler) CreateCourt(c *gin.Context) {
	var input courtRequest
	if request.BindJSON(c, &input) {
		return
	}

	actorID, err := h.tenantActorID(c)
	if err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
		)
		return
	}

	result, err := h.catalog.CreateCourt(catalogTenantContext(c), &catalogV1.CreateCourtRequest{
		ActorUserId:         actorID,
		VenueId:             c.Param("venue_id"),
		Name:                input.Name,
		SlotIntervalMinutes: input.SlotIntervalMinutes,
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusCreated,
		exception.CourtErrors,
	)
}

func (h *CatalogHandler) UpdateCourt(c *gin.Context) {
	var input courtRequest
	if request.BindJSON(c, &input) {
		return
	}

	actorID, err := h.tenantActorID(c)
	if err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
		)
		return
	}

	result, err := h.catalog.UpdateCourt(catalogTenantContext(c), &catalogV1.UpdateCourtRequest{
		ActorUserId:         actorID,
		CourtId:             c.Param("court_id"),
		Name:                input.Name,
		SlotIntervalMinutes: input.SlotIntervalMinutes,
		Status:              input.Status,
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.CourtErrors,
	)
}

func (h *CatalogHandler) ArchiveCourt(c *gin.Context) {
	actorID, err := h.tenantActorID(c)
	if err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
		)
		return
	}

	result, err := h.catalog.ArchiveCourt(catalogTenantContext(c), &catalogV1.ArchiveCourtRequest{
		ActorUserId: actorID,
		CourtId:     c.Param("court_id"),
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusNoContent,
		exception.CourtErrors,
	)
}

func (h *CatalogHandler) tenantActorID(c *gin.Context) (string, error) {
	user, err := h.identity.GetCurrentUser(request.RPCContext(c, platformService.UserGateway), &identityV1.GetCurrentUserRequest{AccessToken: request.BearerToken(c)})
	if err != nil {
		return "", err
	}

	if user.GetId() == "" {
		return "", exception.ErrGRPCTenantUserRequired
	}

	return user.GetId(), nil
}

// catalogTenantContext forwards the caller token for service-side verification.
func catalogTenantContext(c *gin.Context) context.Context {
	ctx := request.RPCContext(c, platformService.UserGateway)
	return metadata.AppendToOutgoingContext(
		ctx,
		"authorization",
		"Bearer "+request.BearerToken(c),
	)
}
