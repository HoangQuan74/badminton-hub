package test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/gateways/user/transport/http/handler"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type identityClient struct {
	identityV1.IdentityAuthServiceClient
}

func (identityClient) GetCurrentUser(_ context.Context, request *identityV1.GetCurrentUserRequest, _ ...grpc.CallOption) (*identityV1.UserResponse, error) {
	return &identityV1.UserResponse{Id: "verified-user"}, nil
}

type catalogClient struct {
	catalogV1.CatalogServiceClient
	actor         string
	authorization []string
}

func (c *catalogClient) ListTenantVenues(ctx context.Context, request *catalogV1.ListTenantVenuesRequest, _ ...grpc.CallOption) (*catalogV1.ListVenuesResponse, error) {
	c.actor = request.GetActorUserId()
	md, _ := metadata.FromOutgoingContext(ctx)
	c.authorization = md.Get("authorization")
	return &catalogV1.ListVenuesResponse{}, nil
}

func TestTenantGatewayForwardsVerifiedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	catalog := &catalogClient{}
	handler.RegisterCatalog(router, identityClient{}, catalog)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tenant/venues", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if catalog.actor != "verified-user" {
		t.Fatalf("actor = %q", catalog.actor)
	}
	if len(catalog.authorization) != 1 || catalog.authorization[0] != "Bearer test-token" {
		t.Fatalf("authorization metadata = %v", catalog.authorization)
	}
}

func TestTenantGatewayRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	catalog := &catalogClient{}
	handler.RegisterCatalog(router, identityClient{}, catalog)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/tenant/venues", nil))
	if response.Code != http.StatusUnauthorized || catalog.actor != "" {
		t.Fatalf("status = %d, actor = %q", response.Code, catalog.actor)
	}
}

type imageCatalogClient struct {
	catalogV1.CatalogServiceClient
	venue   string
	content []byte
	auth    []string
}

func (c *imageCatalogClient) UploadVenueImage(ctx context.Context, request *catalogV1.UploadVenueImageRequest, _ ...grpc.CallOption) (*catalogV1.VenueImageResponse, error) {
	c.venue = request.GetVenueId()
	c.content = request.GetContent()
	metadata, _ := metadata.FromOutgoingContext(ctx)
	c.auth = metadata.Get("authorization")
	return &catalogV1.VenueImageResponse{Image: &catalogV1.VenueImage{Id: "image-id", VenueId: c.venue, ContentType: "image/png", SizeBytes: int32(len(c.content))}}, nil
}

func TestTenantImageUploadForwardsFileAndToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	catalog := &imageCatalogClient{}
	handler.RegisterCatalog(router, identityClient{}, catalog)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", "venue.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("image-content")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tenant/venues/venue-id/images", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer tenant-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if catalog.venue != "venue-id" || string(catalog.content) != "image-content" || len(catalog.auth) != 1 || catalog.auth[0] != "Bearer tenant-token" {
		t.Fatalf("catalog = %+v", catalog)
	}
	if !strings.Contains(response.Body.String(), "/api/v1/tenant/venue-images/image-id") {
		t.Fatalf("missing tenant preview url: %s", response.Body.String())
	}
}

func TestTenantImageUploadRejectsOversizedFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	catalog := &imageCatalogClient{}
	handler.RegisterCatalog(router, identityClient{}, catalog)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", "large.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(make([]byte, 2*1024*1024+1)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tenant/venues/venue-id/images", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer tenant-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || catalog.content != nil {
		t.Fatalf("status = %d, forwarded = %d", response.Code, len(catalog.content))
	}
}
