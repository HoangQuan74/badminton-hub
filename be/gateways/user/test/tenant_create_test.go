package test

import (
	"context"
	"encoding/json"
	handler "github.com/kainguyen/badminton-hub/be/gateways/user/transport/http/handler"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/gateways/user/transport/exception"
	grpcMetadata "github.com/kainguyen/badminton-hub/be/platform/grpc/metadata"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type tenantCreateClient struct {
	identityV1.IdentityTenantServiceClient
	request   *identityV1.TenantCreateOrganizationRequest
	result    *identityV1.TenantOrganizationResponse
	err       error
	wait      bool
	requestID string
}

func (c *tenantCreateClient) CreateOrganization(ctx context.Context, request *identityV1.TenantCreateOrganizationRequest, _ ...grpc.CallOption) (*identityV1.TenantOrganizationResponse, error) {
	c.request = request
	metadata, _ := grpcMetadata.FromContext(ctx)
	c.requestID = metadata.RequestID
	if c.wait {
		<-ctx.Done()
		return nil, status.FromContextError(ctx.Err()).Err()
	}
	return c.result, c.err
}

func tenantCreateResult() *identityV1.TenantOrganizationResponse {
	return &identityV1.TenantOrganizationResponse{
		User:                    &identityV1.UserResponse{Id: "00000000-0000-4000-8000-000000000003", Email: "owner@example.com", FullName: "Owner", Status: "active", CreatedAt: "2026-11-11T03:00:00Z", Roles: []string{"tenant_owner"}},
		Organization:            &identityV1.TenantOrganization{Id: "10000000-0000-4000-8000-000000000001", Name: "Smash Center", Slug: "smash-center", Status: "PENDING_VERIFICATION", CreatedAt: "2026-11-11T03:00:00Z", UpdatedAt: "2026-11-11T03:00:00Z"},
		Membership:              &identityV1.TenantStaffResponse{OrganizationId: "10000000-0000-4000-8000-000000000001", UserId: "00000000-0000-4000-8000-000000000003", Role: "tenant_owner", Status: "ACTIVE"},
		IdempotencyKeyExpiresAt: "2026-11-12T03:00:00Z",
	}
}

func callTenantCreate(t *testing.T, client *tenantCreateClient, body string, headers map[string]string, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterTenant(router, client, 10*time.Millisecond)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tenant"+path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "84f6cb1a-27b1-4e9b-9a14-95e36cfbbac6")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing safe response metadata")
	}
	return response
}

func TestTenantCreateGatewaySuccessAndReplayHeaders(t *testing.T) {
	for _, replayed := range []bool{false, true} {
		client := &tenantCreateClient{result: tenantCreateResult()}
		client.result.IdempotencyReplayed = replayed
		response := callTenantCreate(t, client, `{"name":"Smash Center","slug":"smash-center"}`, nil, "")
		if response.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		if client.request.GetAccessToken() != "token" || client.request.GetIdempotencyKey() == "" || client.requestID != response.Header().Get("X-Request-ID") {
			t.Fatal("request context/input was not forwarded")
		}
		var body map[string]json.RawMessage
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body) != 3 || body["idempotency_replayed"] != nil {
			t.Fatal("internal metadata leaked into body")
		}
		var user map[string]any
		if err := json.Unmarshal(body["user"], &user); err != nil {
			t.Fatal(err)
		}
		if user["phone"] != "" || user["is_admin"] != false {
			t.Fatalf("required default values omitted: %+v", user)
		}
		wantReplay := "false"
		if replayed {
			wantReplay = "true"
		}
		if response.Header().Get("Idempotency-Replayed") != wantReplay || response.Header().Get("Location") == "" || response.Header().Get("Idempotency-Key-Expires-At") == "" {
			t.Fatal("missing replay/location headers")
		}
	}
}

func TestTenantCreateGatewayRejectsMalformedInputBeforeRPC(t *testing.T) {
	valid := `{"name":"Smash Center","slug":"smash-center"}`
	for _, test := range []struct {
		name, body, path string
		headers          map[string]string
		status           int
	}{
		{"missing bearer", valid, "", map[string]string{"Authorization": ""}, 401},
		{"wrong bearer", valid, "", map[string]string{"Authorization": "Basic token"}, 401},
		{"missing key", valid, "", map[string]string{"Idempotency-Key": ""}, 400},
		{"invalid key", valid, "", map[string]string{"Idempotency-Key": "84f6cb1a-27b1-1e9b-9a14-95e36cfbbac6"}, 400},
		{"unknown field", `{"name":"Smash","slug":"smash","is_admin":true}`, "", nil, 400},
		{"invalid json", `{`, "", nil, 400},
		{"multiple objects", valid + valid, "", nil, 400},
		{"oversize", strings.Repeat(" ", 4097), "", nil, 413},
		{"wrong type", valid, "", map[string]string{"Content-Type": "text/plain"}, 415},
		{"wrong charset", valid, "", map[string]string{"Content-Type": "application/json;charset=latin1"}, 415},
		{"invalid language", valid, "?lang=fr", nil, 400},
		{"empty language", valid, "?lang=", nil, 400},
		{"duplicate language", valid, "?lang=vi&lang=en", nil, 400},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := &tenantCreateClient{}
			response := callTenantCreate(t, client, test.body, test.headers, test.path)
			if response.Code != test.status || client.request != nil {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestTenantCreateGatewayMapsStructuredErrorsWithoutLeakingMessages(t *testing.T) {
	for _, test := range []struct {
		code   codes.Code
		reason string
		want   int
	}{
		{codes.AlreadyExists, "ORGANIZATION_SLUG_ALREADY_EXISTS", 409},
		{codes.AlreadyExists, "IDEMPOTENCY_KEY_CONFLICT", 409},
		{codes.FailedPrecondition, "TENANT_ROLE_CONFLICT", 409},
		{codes.InvalidArgument, "INVALID_REQUEST", 400},
		{codes.Unauthenticated, "UNAUTHORIZED", 401},
		{codes.PermissionDenied, "FORBIDDEN", 403},
		{codes.Unavailable, "DEPENDENCY_UNAVAILABLE", 502},
		{codes.DeadlineExceeded, "DEPENDENCY_TIMEOUT", 504},
		{codes.Internal, "INTERNAL_ERROR", 500},
		{codes.AlreadyExists, "UNRECOGNIZED_REASON", 500},
	} {
		value, err := status.New(test.code, "private database/token data").WithDetails(
			&errdetails.ErrorInfo{Reason: test.reason, Domain: "badminton.identity.tenant"},
			&errdetails.BadRequest{FieldViolations: []*errdetails.BadRequest_FieldViolation{{Field: "slug", Description: "SLUG_INVALID"}, {Field: "password", Description: "private data"}}},
		)
		if err != nil {
			t.Fatal(err)
		}
		response := callTenantCreate(t, &tenantCreateClient{err: value.Err()}, `{"name":"Smash","slug":"smash"}`, map[string]string{"Accept-Language": "vi"}, "?lang=en")
		if response.Code != test.want || strings.Contains(response.Body.String(), "private") {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		if test.reason != "UNRECOGNIZED_REASON" && !strings.Contains(response.Body.String(), test.reason) {
			t.Fatalf("reason lost: %s", response.Body.String())
		}
		if strings.Contains(response.Body.String(), "Vui lòng") {
			t.Fatal("query language did not override header")
		}
	}
	response := callTenantCreate(t, &tenantCreateClient{wait: true}, `{"name":"Smash","slug":"smash"}`, nil, "")
	if response.Code != 504 {
		t.Fatalf("timeout status=%d", response.Code)
	}
}

func TestTenantCreateLocalizationFallback(t *testing.T) {
	for _, test := range []struct {
		query, header, want string
		valid               bool
	}{
		{"vi", "en", "vi", true}, {"en", "vi", "en", true}, {"", "vi-VN, en;q=0.5", "vi", true},
		{"", "fr", "en", true}, {"", "invalid;header", "en", true}, {"", "", "en", true}, {"fr", "vi", "en", false},
	} {
		got, valid := exception.TenantLanguage(test.query, test.header)
		if got != test.want || valid != test.valid {
			t.Fatalf("language=%s valid=%v", got, valid)
		}
	}
	if exception.TenantMessage("UNAUTHORIZED", "vi") == exception.TenantMessage("UNAUTHORIZED", "en") {
		t.Fatal("Vietnamese catalog missing")
	}
	if exception.TenantMessage("UNAUTHORIZED", "fr") == "" || exception.TenantMessage("unknown", "en") == "" {
		t.Fatal("fallback missing")
	}
}
