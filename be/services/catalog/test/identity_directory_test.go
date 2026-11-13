package test

import (
	"context"
	"errors"
	"testing"
	"time"

	identityAdapter "github.com/kainguyen/badminton-hub/be/services/catalog/internal/adapter/identity"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/constant"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	catalogException "github.com/kainguyen/badminton-hub/be/services/catalog/internal/exception"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type directoryClientStub struct {
	identityV1.IdentityVenueServiceClient
	context context.Context
	request *identityV1.TenantAuthorizeVenueRequest
	err     error
	calls   int
}

func (s *directoryClientStub) ListPublicVenues(ctx context.Context, _ *identityV1.TenantListPublicVenuesRequest, _ ...grpc.CallOption) (*identityV1.TenantVenuesResponse, error) {
	s.context = ctx
	s.calls++
	return &identityV1.TenantVenuesResponse{}, s.err
}

func (s *directoryClientStub) AuthorizeVenue(ctx context.Context, request *identityV1.TenantAuthorizeVenueRequest, _ ...grpc.CallOption) (*identityV1.TenantVenueResponse, error) {
	s.context = ctx
	s.request = request
	s.calls++
	return &identityV1.TenantVenueResponse{Venue: &identityV1.TenantVenue{Id: request.GetVenueId()}}, s.err
}

func (s *directoryClientStub) CreateVenue(ctx context.Context, _ *identityV1.TenantCreateVenueRequest, _ ...grpc.CallOption) (*identityV1.TenantVenueResponse, error) {
	s.context = ctx
	s.calls++
	return nil, s.err
}

func TestDirectoryBoundsDeadlineAndReleasesContext(t *testing.T) {
	client := &directoryClientStub{}
	directory := identityAdapter.NewDirectory(client)
	started := time.Now()
	if _, err := directory.ListPublicVenues(context.Background()); err != nil {
		t.Fatal(err)
	}
	deadline, ok := client.context.Deadline()
	if !ok || deadline.After(time.Now().Add(constant.IdentityRPCTimeout)) || deadline.Before(started) {
		t.Fatal("Identity call must use a bounded request deadline")
	}
	if !errors.Is(client.context.Err(), context.Canceled) {
		t.Fatal("completed Identity call must release its request context")
	}

	// A shorter caller deadline must take precedence over the adapter timeout.
	parent, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	parentDeadline, _ := parent.Deadline()
	if _, err := directory.ListPublicVenues(parent); err != nil {
		t.Fatal(err)
	}
	deadline, _ = client.context.Deadline()
	if !deadline.Equal(parentDeadline) {
		t.Fatal("adapter extended the caller's deadline")
	}
}

func TestDirectoryRequiresTokenAndPreservesOwnerScope(t *testing.T) {
	client := &directoryClientStub{}
	directory := identityAdapter.NewDirectory(client)
	if _, err := directory.AuthorizeVenue(context.Background(), "venue-id", true); !errors.Is(err, catalogException.ErrUnauthenticated) {
		t.Fatalf("missing token: %v", err)
	}
	if client.calls != 0 {
		t.Fatal("unauthenticated request reached Identity")
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer test-token"))
	venue, err := directory.AuthorizeVenue(ctx, "venue-id", true)
	if err != nil || venue.ID != "venue-id" {
		t.Fatalf("authorize venue: %v", err)
	}
	if client.request.GetAccessToken() != "test-token" || !client.request.GetRequireOwner() {
		t.Fatal("adapter changed the verified token or requested owner scope")
	}
}

func TestDirectoryPreservesDependencyAndConflictErrors(t *testing.T) {
	client := &directoryClientStub{err: status.Error(codes.Unavailable, "Identity unavailable")}
	directory := identityAdapter.NewDirectory(client)
	if _, err := directory.ListPublicVenues(context.Background()); !errors.Is(err, catalogException.ErrIdentityUnavailable) {
		t.Fatalf("dependency error: %v", err)
	}

	client.err = status.Error(codes.AlreadyExists, "slug taken")
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer test-token"))
	if _, err := directory.CreateVenue(ctx, catalog.CreateVenueInput{}); !errors.Is(err, catalogException.ErrVenueSlugConflict) {
		t.Fatalf("venue slug conflict: %v", err)
	}
}
