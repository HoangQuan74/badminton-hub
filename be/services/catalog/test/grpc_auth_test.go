package test

import (
	"context"
	"errors"
	"testing"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	cataloggrpc "github.com/kainguyen/badminton-hub/be/services/catalog/internal/transport/grpc/catalog"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type identityStub struct {
	identityV1.IdentityAuthServiceClient
	userID string
	roles  []string
	err    error
	token  string
	portal string
}

func (s *identityStub) GetCurrentUser(_ context.Context, request *identityV1.GetCurrentUserRequest, _ ...grpc.CallOption) (*identityV1.UserResponse, error) {
	s.token = request.GetAccessToken()
	s.portal = request.GetPortal()
	if s.err != nil {
		return nil, s.err
	}
	return &identityV1.UserResponse{Id: s.userID, Roles: s.roles}, nil
}

type actorRepository struct {
	repositoryStub
	actor string
	err   error
}

func (r *actorRepository) ListTenantVenues(_ context.Context, actor string) ([]model.Venue, error) {
	r.actor = actor
	return nil, r.err
}

func TestTenantGRPCVerifiesTokenAndIgnoresSuppliedActor(t *testing.T) {
	repository := &actorRepository{}
	identity := &identityStub{userID: "verified-user"}
	server := cataloggrpc.NewServer(catalog.NewService(repository), identity)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer valid-token"))
	_, err := server.ListTenantVenues(ctx, &catalogV1.ListTenantVenuesRequest{ActorUserId: "forged-user"})
	if err != nil {
		t.Fatalf("ListTenantVenues() error = %v", err)
	}
	if repository.actor != "verified-user" {
		t.Fatalf("actor = %q, want verified-user", repository.actor)
	}
	if identity.token != "valid-token" || identity.portal != "tenant" {
		t.Fatalf("Identity received token %q and portal %q", identity.token, identity.portal)
	}
}

func TestTenantGRPCFailsClosed(t *testing.T) {
	tests := []struct {
		name        string
		header      []string
		identityErr error
		identityID  string
		want        codes.Code
	}{
		{name: "missing token", want: codes.Unauthenticated},
		{name: "invalid scheme", header: []string{"Basic value"}, want: codes.Unauthenticated},
		{name: "multiple tokens", header: []string{"Bearer one", "Bearer two"}, want: codes.Unauthenticated},
		{name: "invalid token", header: []string{"Bearer invalid"}, identityErr: status.Error(codes.Unauthenticated, "invalid credentials"), want: codes.Unauthenticated},
		{name: "identity unavailable", header: []string{"Bearer valid"}, identityErr: status.Error(codes.Unavailable, "unavailable"), want: codes.Unavailable},
		{name: "missing user", header: []string{"Bearer valid"}, want: codes.Unauthenticated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &actorRepository{}
			identity := &identityStub{userID: tt.identityID, err: tt.identityErr}
			server := cataloggrpc.NewServer(catalog.NewService(repository), identity)
			ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{"authorization": tt.header})
			_, err := server.ListTenantVenues(ctx, &catalogV1.ListTenantVenuesRequest{ActorUserId: "forged-user"})
			if status.Code(err) != tt.want {
				t.Fatalf("code = %v, want %v", status.Code(err), tt.want)
			}
			if repository.actor != "" {
				t.Fatalf("repository called with actor %q", repository.actor)
			}
		})
	}
}

func TestGRPCDoesNotExposeDatabaseError(t *testing.T) {
	repository := &actorRepository{err: errors.New("database credentials or query")}
	server := cataloggrpc.NewServer(catalog.NewService(repository), &identityStub{userID: "verified-user"})
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer valid"))
	_, err := server.ListTenantVenues(ctx, &catalogV1.ListTenantVenuesRequest{})
	if status.Code(err) != codes.Internal || status.Convert(err).Message() != "catalog operation failed" {
		t.Fatalf("error = %v", err)
	}
}

func TestCreateOrganizationRequiresVerifiedOwnerRole(t *testing.T) {
	repository := &repositoryStub{}
	identity := &identityStub{userID: "verified-user", roles: []string{"tenant_staff"}}
	server := cataloggrpc.NewServer(catalog.NewService(repository), identity)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer valid-token"))
	_, err := server.CreateOrganization(ctx, &catalogV1.CreateOrganizationRequest{Name: "Test", Slug: "test"})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("code = %v, want permission denied", status.Code(err))
	}
}
