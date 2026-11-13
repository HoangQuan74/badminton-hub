package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/gateways/user/transport/exception"
	tenantRequest "github.com/kainguyen/badminton-hub/be/gateways/user/transport/http/request"
	httpContext "github.com/kainguyen/badminton-hub/be/platform/http/context"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type TenantHandler struct {
	client  identityV1.IdentityTenantServiceClient
	timeout time.Duration
}

func RegisterTenant(engine *gin.Engine, client identityV1.IdentityTenantServiceClient, timeout time.Duration) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	handler := &TenantHandler{
		client:  client,
		timeout: timeout,
	}
	// Register tenant CRUD together under the owning feature.
	engine.POST("/api/v1/tenant", handler.Create)
	group := engine.Group("/api/v1/tenant")
	group.Use(requireTenantBearer)
	group.GET("", handler.ListOrganizations)
	group.GET("/:organization_id", handler.GetOrganization)
	group.PATCH("/:organization_id", handler.UpdateOrganization)
	group.DELETE("/:organization_id", handler.ArchiveOrganization)

}

func (h *TenantHandler) Create(c *gin.Context) {
	// Ensure safe request metadata and uncached responses even in standalone routing.
	metadata, ok := httpContext.RequestMetadataFromContext(c.Request.Context())
	if !ok {
		requestID := uuid.NewString()
		metadata = httpContext.RequestMetadata{
			RequestID:     requestID,
			CorrelationID: requestID,
		}
		c.Request = c.Request.WithContext(httpContext.WithRequestMetadata(c.Request.Context(), metadata))
	}

	c.Header("X-Request-ID", metadata.RequestID)
	c.Header("Cache-Control", "no-store")
	lang, validLanguage := exception.TenantLanguage(c.Query("lang"), c.GetHeader("Accept-Language"))
	if values, present := c.Request.URL.Query()["lang"]; present && (len(values) != 1 || values[0] == "") {
		lang, validLanguage = "en", false
	}

	if !validLanguage {
		exception.WriteTenantError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			lang,
			nil,
		)
		return
	}

	// Bind only one bearer token and one UUID-v4 idempotency key.
	authorization := strings.Fields(c.GetHeader("Authorization"))
	if len(c.Request.Header.Values("Authorization")) != 1 || len(authorization) != 2 || !strings.EqualFold(authorization[0], "Bearer") {
		exception.WriteTenantError(
			c,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			lang,
			nil,
		)
		return
	}

	key := c.GetHeader("Idempotency-Key")
	parsedKey, err := uuid.Parse(key)
	if len(c.Request.Header.Values("Idempotency-Key")) != 1 || len(key) != 36 || err != nil || parsedKey.Version() != 4 || parsedKey.Variant() != uuid.RFC4122 {
		fieldCode := "UUID_INVALID"
		if key == "" {
			fieldCode = "REQUIRED"
		}

		exception.WriteTenantError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			lang,
			[]exception.TenantFieldError{{
				Field: "idempotency_key",
				Code:  fieldCode,
			}},
		)
		return
	}

	// Decode a bounded JSON body and reject unknown fields or trailing objects.
	mediaType, parameters, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" || (parameters["charset"] != "" && !strings.EqualFold(parameters["charset"], "utf-8")) {
		exception.WriteTenantError(
			c,
			http.StatusUnsupportedMediaType,
			"UNSUPPORTED_MEDIA_TYPE",
			lang,
			nil,
		)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		4096,
	))
	if err != nil {
		var oversized *http.MaxBytesError
		if errors.As(err, &oversized) {
			exception.WriteTenantError(
				c,
				http.StatusRequestEntityTooLarge,
				"PAYLOAD_TOO_LARGE",
				lang,
				nil,
			)
		} else {
			exception.WriteTenantError(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				lang,
				nil,
			)
		}

		return
	}

	var input tenantRequest.CreateTenantRequest
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if !utf8.Valid(body) || decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF {
		exception.WriteTenantError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			lang,
			nil,
		)
		return
	}

	// Propagate request context and a bounded deadline to the owning service.
	ctx, cancel := context.WithTimeout(request.RPCContext(c, platformService.UserGateway), h.timeout)
	defer cancel()
	result, err := h.client.CreateOrganization(ctx, &identityV1.TenantCreateOrganizationRequest{
		AccessToken:    authorization[1],
		Name:           input.Name,
		Slug:           input.Slug,
		IdempotencyKey: key,
		ContactName:    input.ContactName,
		ContactEmail:   input.ContactEmail,
		ContactPhone:   input.ContactPhone,
		Note:           input.Note,
	})
	if err != nil {
		exception.WriteTenantGRPCError(
			c,
			err,
			lang,
		)
		return
	}

	if result.GetUser() == nil || result.GetOrganization() == nil || result.GetMembership() == nil {
		exception.WriteTenantError(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			lang,
			nil,
		)
		return
	}

	// Expose replay metadata as headers, never as extra body fields.
	c.Header("Location", "/api/v1/tenant/"+url.PathEscape(result.GetOrganization().GetId()))
	c.Header("Idempotency-Replayed", "false")
	if result.GetIdempotencyReplayed() {
		c.Header("Idempotency-Replayed", "true")
	}

	c.Header("Idempotency-Key-Expires-At", result.GetIdempotencyKeyExpiresAt())
	c.JSON(http.StatusCreated, tenantResponse(result))
}

func (h *TenantHandler) ListOrganizations(c *gin.Context) {
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

	result, err := h.client.ListOrganizations(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantListOrganizationsRequest{AccessToken: request.BearerToken(c)})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.OrganizationErrors,
	)
}

func (h *TenantHandler) GetOrganization(c *gin.Context) {
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

	result, err := h.client.GetOrganization(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantGetOrganizationRequest{
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

func (h *TenantHandler) UpdateOrganization(c *gin.Context) {
	var input tenantRequest.UpdateTenantRequest
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

	result, err := h.client.UpdateOrganization(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantUpdateOrganizationRequest{
		AccessToken:    request.BearerToken(c),
		OrganizationId: c.Param("organization_id"),
		Name:           input.Name,
		Slug:           input.Slug,
	})
	exception.WriteGRPC(
		c,
		result,
		err,
		http.StatusOK,
		exception.OrganizationErrors,
	)
}

func (h *TenantHandler) ArchiveOrganization(c *gin.Context) {
	if h.client == nil {
		exception.WriteGRPC(
			c,
			nil,
			exception.ErrGRPCIdentityUnavailable,
			http.StatusNoContent,
			exception.OrganizationErrors,
		)
		return
	}

	_, err := h.client.ArchiveOrganization(request.RPCContext(c, platformService.UserGateway), &identityV1.TenantArchiveOrganizationRequest{
		AccessToken:    request.BearerToken(c),
		OrganizationId: c.Param("organization_id"),
	})
	exception.WriteGRPC(
		c,
		nil,
		err,
		http.StatusNoContent,
		exception.OrganizationErrors,
	)
}

// Explicit HTTP fields preserve required empty strings/false values that proto
// structs would omit under encoding/json. Internal replay metadata stays out.
type tenantUserResponse struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	FullName  string   `json:"full_name"`
	Phone     string   `json:"phone"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
	Roles     []string `json:"roles"`
	IsAdmin   bool     `json:"is_admin"`
}

type tenantOrganizationResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type tenantMembershipResponse struct {
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	Role           string `json:"role"`
	Status         string `json:"status"`
}

type tenantCreateResponse struct {
	User         tenantUserResponse         `json:"user"`
	Organization tenantOrganizationResponse `json:"organization"`
	Membership   tenantMembershipResponse   `json:"membership"`
}

func tenantResponse(value *identityV1.TenantOrganizationResponse) tenantCreateResponse {
	user, organization, membership := value.GetUser(), value.GetOrganization(), value.GetMembership()
	return tenantCreateResponse{
		User: tenantUserResponse{
			ID:        user.GetId(),
			Email:     user.GetEmail(),
			FullName:  user.GetFullName(),
			Phone:     user.GetPhone(),
			Status:    user.GetStatus(),
			CreatedAt: user.GetCreatedAt(),
			Roles:     user.GetRoles(),
			IsAdmin:   user.GetIsAdmin(),
		},
		Organization: tenantOrganizationResponse{
			ID:        organization.GetId(),
			Name:      organization.GetName(),
			Slug:      organization.GetSlug(),
			Status:    organization.GetStatus(),
			CreatedAt: organization.GetCreatedAt(),
			UpdatedAt: organization.GetUpdatedAt(),
		},
		Membership: tenantMembershipResponse{
			OrganizationID: membership.GetOrganizationId(),
			UserID:         membership.GetUserId(),
			Role:           membership.GetRole(),
			Status:         membership.GetStatus(),
		},
	}
}
