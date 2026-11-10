package test

import (
	"bytes"
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	s3store "github.com/kainguyen/badminton-hub/be/services/catalog/internal/storage/objectstore/s3"
	catalogpostgres "github.com/kainguyen/badminton-hub/be/services/catalog/internal/storage/postgres/repository/catalog"
	cataloggrpc "github.com/kainguyen/badminton-hub/be/services/catalog/internal/transport/grpc/catalog"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

// Both services must be disposable: this test inserts venue and image records.
func TestImagePostgresS3Integration(t *testing.T) {
	databaseURL, endpoint := os.Getenv("TEST_CATALOG_DATABASE_URL"), os.Getenv("TEST_CATALOG_S3_ENDPOINT")
	if databaseURL == "" || endpoint == "" {
		t.Skip("dedicated Catalog PostgreSQL and S3 test services are required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	storage := catalogpostgres.New(pool)
	directory := newDirectoryStub()
	repository := catalog.NewScopedRepository(storage, directory)
	store, err := s3store.New(ctx, s3store.Config{Bucket: "catalog-test", Region: "us-east-1", Endpoint: endpoint, AccessKey: "catalogtest", SecretKey: "catalogtest"})
	if err != nil {
		t.Fatal(err)
	}
	service := catalog.NewImageService(repository, store)
	owner := "00000000-0000-4000-8000-000000000003"
	staff := "00000000-0000-4000-8000-000000000004"
	outsider := "00000000-0000-4000-8000-000000000005"
	published := "20000000-0000-4000-8000-000000000001"
	content := samplePNG(t)
	ownerCtx := metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer owner"))
	staffCtx := metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer staff"))
	outsiderCtx := metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer outsider"))
	if _, err := service.Upload(staffCtx, staff, published, content); !errors.Is(err, catalog.ErrForbidden) {
		t.Fatalf("staff upload error = %v", err)
	}
	uploaded, err := service.Upload(ownerCtx, owner, published, content)
	if err != nil {
		t.Fatalf("upload published image: %v", err)
	}
	item, actual, err := service.GetPublic(ctx, uploaded.ID)
	if err != nil || item.ContentType != "image/png" || !bytes.Equal(actual, content) {
		t.Fatalf("get public image: %#v, %v", item, err)
	}
	if _, _, err := service.GetTenant(outsiderCtx, outsider, uploaded.ID); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("outsider get = %v", err)
	}
	venue, err := repository.CreateVenue(ownerCtx, owner, catalog.CreateVenueInput{OrganizationID: "10000000-0000-4000-8000-000000000001", Name: "Draft with image", Slug: "image-" + uuid.NewString(), Timezone: "Asia/Ho_Chi_Minh"})
	if err != nil {
		t.Fatal(err)
	}
	draftImage, err := service.Upload(ownerCtx, owner, venue.ID, content)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.GetPublic(ctx, draftImage.ID); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("draft public image = %v", err)
	}
	if _, actual, err := service.GetTenant(staffCtx, staff, draftImage.ID); err != nil || !bytes.Equal(actual, content) {
		t.Fatalf("staff draft preview: %v", err)
	}
	if _, err := repository.ArchiveVenue(ownerCtx, owner, venue.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.GetTenant(ownerCtx, owner, draftImage.ID); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("archived preview = %v", err)
	}

	// Exercise actual gRPC serialization and authorization metadata with real PostgreSQL and S3.
	listener := bufconn.Listen(4 * 1024 * 1024)
	grpcServer := grpc.NewServer()
	catalogV1.RegisterCatalogServiceServer(grpcServer, cataloggrpc.NewServer(catalog.NewService(repository), &identityStub{userID: owner, roles: []string{"tenant_owner"}}, service))
	go func() { _ = grpcServer.Serve(listener) }()
	defer grpcServer.Stop()
	connection, err := grpc.NewClient("passthrough:///catalog-test", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	client := catalogV1.NewCatalogServiceClient(connection)
	authCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer owner"))
	response, err := client.UploadVenueImage(authCtx, &catalogV1.UploadVenueImageRequest{VenueId: published, Content: content})
	if err != nil {
		t.Fatalf("gRPC upload: %v", err)
	}
	read, err := client.GetPublicVenueImage(ctx, &catalogV1.GetVenueImageRequest{ImageId: response.GetImage().GetId()})
	if err != nil || read.GetContentType() != "image/png" || !bytes.Equal(read.GetContent(), content) {
		t.Fatalf("gRPC public image: %v", err)
	}
}
