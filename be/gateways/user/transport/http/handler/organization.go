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

type createOrganizationRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}
type updateOrganizationRequest struct {
	Name *string `json:"name"`
	Slug *string `json:"slug"`
}

var organizationErrors = exception.ErrorConfig{NotFoundCode: "ORGANIZATION_NOT_FOUND", AlreadyExistsCode: "ORGANIZATION_SLUG_ALREADY_EXISTS", FailedPreconditionCode: "ORGANIZATION_STATE_CONFLICT", UnavailableCode: "DEPENDENCY_UNAVAILABLE"}

func (h *CatalogHandler) ListOrganizations(c *gin.Context) {
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(c, nil, err, http.StatusOK, organizationErrors)
		return
	}
	if h.directory == nil {
		exception.WriteGRPC(c, nil, status.Error(codes.Unavailable, "identity unavailable"), http.StatusOK, organizationErrors)
		return
	}
	result, err := h.directory.ListOrganizations(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantListOrganizationsRequest{AccessToken: request.BearerToken(c)})
	exception.WriteGRPC(c, result, err, http.StatusOK, organizationErrors)
}

func (h *CatalogHandler) GetOrganization(c *gin.Context) {
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(c, nil, err, http.StatusOK, organizationErrors)
		return
	}
	if h.directory == nil {
		exception.WriteGRPC(c, nil, status.Error(codes.Unavailable, "identity unavailable"), http.StatusOK, organizationErrors)
		return
	}
	result, err := h.directory.GetOrganization(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantGetOrganizationRequest{AccessToken: request.BearerToken(c), OrganizationId: c.Param("organization_id")})
	exception.WriteGRPC(c, result, err, http.StatusOK, organizationErrors)
}

func (h *CatalogHandler) CreateOrganization(c *gin.Context) {
	var input createOrganizationRequest
	if request.BindJSON(c, &input) {
		return
	}
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(c, nil, err, http.StatusCreated, organizationErrors)
		return
	}
	if h.directory == nil {
		exception.WriteGRPC(c, nil, status.Error(codes.Unavailable, "identity unavailable"), http.StatusOK, organizationErrors)
		return
	}
	result, err := h.directory.CreateOrganization(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantCreateOrganizationRequest{AccessToken: request.BearerToken(c), Name: input.Name, Slug: input.Slug})
	exception.WriteGRPC(c, result, err, http.StatusCreated, organizationErrors)
}

func (h *CatalogHandler) UpdateOrganization(c *gin.Context) {
	var input updateOrganizationRequest
	if request.BindJSON(c, &input) {
		return
	}
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(c, nil, err, http.StatusOK, organizationErrors)
		return
	}
	if h.directory == nil {
		exception.WriteGRPC(c, nil, status.Error(codes.Unavailable, "identity unavailable"), http.StatusOK, organizationErrors)
		return
	}
	result, err := h.directory.UpdateOrganization(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantUpdateOrganizationRequest{AccessToken: request.BearerToken(c), OrganizationId: c.Param("organization_id"), Name: input.Name, Slug: input.Slug})
	exception.WriteGRPC(c, result, err, http.StatusOK, organizationErrors)
}

func (h *CatalogHandler) ArchiveOrganization(c *gin.Context) {
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(c, nil, err, http.StatusNoContent, organizationErrors)
		return
	}
	if h.directory == nil {
		exception.WriteGRPC(c, nil, status.Error(codes.Unavailable, "identity unavailable"), http.StatusNoContent, organizationErrors)
		return
	}
	_, err := h.directory.ArchiveOrganization(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantArchiveOrganizationRequest{AccessToken: request.BearerToken(c), OrganizationId: c.Param("organization_id")})
	exception.WriteGRPC(c, nil, err, http.StatusNoContent, organizationErrors)
}
