package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/gateways/user/transport/http/handler"
	grpcClient "github.com/kainguyen/badminton-hub/be/platform/grpc/client"
	httpContext "github.com/kainguyen/badminton-hub/be/platform/http/context"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

// Run only against Identity configured with a disposable migrated database.
func TestTenantCreateHTTPThroughIdentityAndPostgres(t *testing.T) {
	target := os.Getenv("TEST_IDENTITY_GRPC_TARGET")
	if target == "" {
		t.Skip("TEST_IDENTITY_GRPC_TARGET must point to disposable Identity")
	}
	connection, err := grpcClient.Dial(context.Background(), target, grpcClient.Options{Block: true, DialTimeout: 5 * time.Second, ServiceIdentity: platformService.UserGateway})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(httpContext.RequestContextHeaders())
	authClient := identityV1.NewIdentityAuthServiceClient(connection)
	handler.RegisterAuth(router, authClient)
	handler.RegisterTenantAuth(router, identityV1.NewIdentityTenantAuthServiceClient(connection))
	handler.RegisterCatalog(router, identityV1.NewIdentityTenantAuthServiceClient(connection), &catalogClient{})
	handler.RegisterVenue(router, identityV1.NewIdentityVenueServiceClient(connection))
	handler.RegisterStaff(router, identityV1.NewIdentityStaffServiceClient(connection))
	handler.RegisterTenant(router, identityV1.NewIdentityTenantServiceClient(connection), 5*time.Second)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	call := func(method, path, token, key string, body any, want int) ([]byte, http.Header) {
		t.Helper()
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequest(method, server.URL+path, bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		if key != "" {
			request.Header.Set("Idempotency-Key", key)
		}
		response, err := server.Client().Do(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != want {
			t.Fatalf("%s %s: status=%d body=%s", method, path, response.StatusCode, data)
		}
		return data, response.Header
	}
	email := uuid.NewString() + "@example.test"
	credentials := map[string]string{"email": email, "password": "tenantCreate-password"}
	registration, _ := call("POST", "/api/v1/auth/register", "", "", map[string]string{"email": email, "password": "tenantCreate-password", "full_name": "HTTP owner"}, 201)
	var auth struct {
		Tokens struct {
			AccessToken string `json:"access_token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(registration, &auth); err != nil || auth.Tokens.AccessToken == "" {
		t.Fatalf("registration=%s err=%v", registration, err)
	}
	token := auth.Tokens.AccessToken
	call("POST", "/api/v1/tenant/auth/login", "", "", credentials, 403)
	call("POST", "/api/v1/tenant", token, "", map[string]string{"name": "Denied", "slug": "denied-" + uuid.NewString()}, 400)
	payload := map[string]string{"name": "HTTP owner organization", "slug": "http-" + uuid.NewString()}
	key := uuid.NewString()
	first, headers := call("POST", "/api/v1/tenant", token, key, payload, 201)
	if headers.Get("Idempotency-Replayed") != "false" || headers.Get("Cache-Control") != "no-store" {
		t.Fatal("missing response guarantees")
	}
	var body struct {
		User struct {
			ID      string   `json:"id"`
			Roles   []string `json:"roles"`
			Phone   *string  `json:"phone"`
			IsAdmin *bool    `json:"is_admin"`
		} `json:"user"`
		Organization struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"organization"`
		Membership struct {
			UserID         string `json:"user_id"`
			OrganizationID string `json:"organization_id"`
			Status         string `json:"status"`
		} `json:"membership"`
	}
	if err := json.Unmarshal(first, &body); err != nil {
		t.Fatal(err)
	}
	if body.Organization.Status != "PENDING_VERIFICATION" || body.Membership.Status != "ACTIVE" || body.User.Phone == nil || body.User.IsAdmin == nil || len(body.User.Roles) != 1 || body.User.Roles[0] != "tenant_owner" {
		t.Fatalf("response=%s", first)
	}
	if body.User.ID != body.Membership.UserID || body.Organization.ID != body.Membership.OrganizationID {
		t.Fatal("membership belongs to wrong actor/organization")
	}
	replay, replayHeaders := call("POST", "/api/v1/tenant", token, key, payload, 201)
	if !bytes.Equal(first, replay) || replayHeaders.Get("Idempotency-Replayed") != "true" || headers.Get("Idempotency-Key-Expires-At") != replayHeaders.Get("Idempotency-Key-Expires-At") {
		t.Fatal("HTTP replay differs")
	}
	if headers.Get("X-Request-ID") == replayHeaders.Get("X-Request-ID") {
		t.Fatal("replay reused original request ID")
	}
	payload["name"] = "Changed"
	conflict, _ := call("POST", "/api/v1/tenant", token, key, payload, 409)
	if !bytes.Contains(conflict, []byte("IDEMPOTENCY_KEY_CONFLICT")) {
		t.Fatalf("conflict=%s", conflict)
	}
	payload["name"] = "HTTP owner organization"
	conflict, _ = call("POST", "/api/v1/tenant", token, uuid.NewString(), payload, 409)
	if !bytes.Contains(conflict, []byte("ORGANIZATION_SLUG_ALREADY_EXISTS")) {
		t.Fatalf("slug conflict=%s", conflict)
	}
	call("GET", "/api/v1/tenant/me", token, "", nil, 200)
	call("GET", "/api/v1/tenant/"+body.Organization.ID, token, "", nil, 200)
	call("POST", "/api/v1/tenant/"+body.Organization.ID+"/venues", token, "", map[string]string{"name": "Pending venue", "slug": "pending-" + uuid.NewString()}, 403)
	call("POST", "/api/v1/auth/login", "", "", credentials, 200)
	call("POST", "/api/v1/tenant/auth/login", "", "", credentials, 200)

	// Verify tenant CRUD and the preserved owner-only organization route.
	detailPath := "/api/v1/tenant/" + body.Organization.ID
	call("GET", detailPath, token, "", nil, 200)
	listing, _ := call("GET", "/api/v1/tenant", token, "", nil, 200)
	if !bytes.Contains(listing, []byte(body.Organization.ID)) {
		t.Fatalf("tenant missing from list: %s", listing)
	}
	call("GET", "/api/v1/tenant", token, "", nil, 200)

}
