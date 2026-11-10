package cataloggrpc

import (
	"context"
	"strings"
	"time"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	catalogV1.UnimplementedCatalogServiceServer
	service  *catalog.Service
	identity identityV1.IdentityAuthServiceClient
	images   *catalog.ImageService
}

func NewServer(service *catalog.Service, identity identityV1.IdentityAuthServiceClient, imageServices ...*catalog.ImageService) *Server {
	server := &Server{service: service, identity: identity}
	if len(imageServices) > 0 {
		server.images = imageServices[0]
	}
	return server
}

func (s *Server) ListPublicVenues(ctx context.Context, request *catalogV1.ListPublicVenuesRequest) (*catalogV1.ListVenuesResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.ListPublicVenuesRequest) (*catalogV1.ListVenuesResponse, error) {
		venues, err := s.service.ListPublicVenues(ctx)
		if err != nil {
			return nil, err
		}
		return &catalogV1.ListVenuesResponse{Items: mapVenues(venues)}, nil
	})
}

func (s *Server) ListPublicVenueCourts(ctx context.Context, request *catalogV1.ListPublicVenueCourtsRequest) (*catalogV1.ListCourtsResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.ListPublicVenueCourtsRequest) (*catalogV1.ListCourtsResponse, error) {
		courts, err := s.service.ListPublicVenueCourts(ctx, request.GetVenueId())
		if err != nil {
			return nil, err
		}
		return &catalogV1.ListCourtsResponse{Items: mapCourts(courts)}, nil
	})
}

func (s *Server) ListTenantVenues(ctx context.Context, request *catalogV1.ListTenantVenuesRequest) (*catalogV1.ListVenuesResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.ListTenantVenuesRequest) (*catalogV1.ListVenuesResponse, error) {
		venues, err := s.service.ListTenantVenues(ctx, actorID)
		if err != nil {
			return nil, err
		}
		return &catalogV1.ListVenuesResponse{Items: mapVenues(venues)}, nil
	})
}

func (s *Server) ListTenantCourts(ctx context.Context, request *catalogV1.ListTenantCourtsRequest) (*catalogV1.ListCourtsResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.ListTenantCourtsRequest) (*catalogV1.ListCourtsResponse, error) {
		courts, err := s.service.ListTenantCourts(ctx, actorID, request.GetVenueId())
		if err != nil {
			return nil, err
		}
		return &catalogV1.ListCourtsResponse{Items: mapCourts(courts)}, nil
	})
}

func (s *Server) CreateCourt(ctx context.Context, request *catalogV1.CreateCourtRequest) (*catalogV1.CourtResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.CreateCourtRequest) (*catalogV1.CourtResponse, error) {
		court, err := s.service.CreateCourt(ctx, actorID, request.GetVenueId(), request.GetName(), request.GetSlotIntervalMinutes())
		if err != nil {
			return nil, err
		}
		return &catalogV1.CourtResponse{Court: mapCourt(court)}, nil
	})
}

func (s *Server) UpdateCourt(ctx context.Context, request *catalogV1.UpdateCourtRequest) (*catalogV1.CourtResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.UpdateCourtRequest) (*catalogV1.CourtResponse, error) {
		court, err := s.service.UpdateCourt(ctx, actorID, request.GetCourtId(), request.GetName(), request.GetSlotIntervalMinutes(), request.GetStatus())
		if err != nil {
			return nil, err
		}
		return &catalogV1.CourtResponse{Court: mapCourt(court)}, nil
	})
}

func (s *Server) ArchiveCourt(ctx context.Context, request *catalogV1.ArchiveCourtRequest) (*catalogV1.ArchiveCourtResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.ArchiveCourtRequest) (*catalogV1.ArchiveCourtResponse, error) {
		court, err := s.service.ArchiveCourt(ctx, actorID, request.GetCourtId())
		if err != nil {
			return nil, err
		}
		return &catalogV1.ArchiveCourtResponse{Court: mapCourt(court)}, nil
	})
}

// tenantActor verifies the bearer token with Identity; request actor IDs are not trusted.
func (s *Server) tenantActor(ctx context.Context) (*identityV1.UserResponse, error) {
	incoming, _ := metadata.FromIncomingContext(ctx)
	values := incoming.Get("authorization")
	if len(values) != 1 {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	scheme, token, found := strings.Cut(values[0], " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	if s.identity == nil {
		return nil, status.Error(codes.Unavailable, "identity unavailable")
	}
	identityCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	user, err := s.identity.GetCurrentUser(identityCtx, &identityV1.GetCurrentUserRequest{AccessToken: strings.TrimSpace(token), Portal: "tenant"}, grpc.WaitForReady(false))
	if err != nil {
		return nil, err
	}
	if user.GetId() == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	return user, nil
}

func (s *Server) tenantActorID(ctx context.Context) (string, error) {
	user, err := s.tenantActor(ctx)
	if err != nil {
		return "", err
	}
	return user.GetId(), nil
}

func (s *Server) tenantOwnerID(ctx context.Context) (string, error) {
	user, err := s.tenantActor(ctx)
	if err != nil {
		return "", err
	}
	for _, role := range user.GetRoles() {
		if role == "tenant_owner" {
			return user.GetId(), nil
		}
	}
	return "", status.Error(codes.PermissionDenied, "tenant owner required")
}
