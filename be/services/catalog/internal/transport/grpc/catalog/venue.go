package cataloggrpc

import (
	"context"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
)

func (s *Server) GetPublicVenue(ctx context.Context, request *catalogV1.GetPublicVenueRequest) (*catalogV1.VenueResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.GetPublicVenueRequest) (*catalogV1.VenueResponse, error) {
		venue, err := s.service.GetPublicVenue(ctx, request.GetSlug())
		if err != nil {
			return nil, err
		}
		return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
	})
}

func (s *Server) GetTenantVenue(ctx context.Context, request *catalogV1.GetTenantVenueRequest) (*catalogV1.VenueResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.GetTenantVenueRequest) (*catalogV1.VenueResponse, error) {
		venue, err := s.service.GetTenantVenue(ctx, actorID, request.GetVenueId())
		if err != nil {
			return nil, err
		}
		return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
	})
}

func (s *Server) CreateVenue(ctx context.Context, request *catalogV1.CreateVenueRequest) (*catalogV1.VenueResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.CreateVenueRequest) (*catalogV1.VenueResponse, error) {
		input := catalog.CreateVenueInput{OrganizationID: request.GetOrganizationId(), Name: request.GetName(), Slug: request.GetSlug(), Address: request.GetAddress(), Timezone: request.GetTimezone()}
		venue, err := s.service.CreateVenue(ctx, actorID, input)
		if err != nil {
			return nil, err
		}
		return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
	})
}

func (s *Server) UpdateVenue(ctx context.Context, request *catalogV1.UpdateVenueRequest) (*catalogV1.VenueResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.UpdateVenueRequest) (*catalogV1.VenueResponse, error) {
		input := catalog.UpdateVenueInput{Name: request.Name, Slug: request.Slug, Address: request.Address, Timezone: request.Timezone}
		venue, err := s.service.UpdateVenue(ctx, actorID, request.GetVenueId(), input)
		if err != nil {
			return nil, err
		}
		return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
	})
}

func (s *Server) ArchiveVenue(ctx context.Context, request *catalogV1.ArchiveVenueRequest) (*catalogV1.VenueResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.ArchiveVenueRequest) (*catalogV1.VenueResponse, error) {
		venue, err := s.service.ArchiveVenue(ctx, actorID, request.GetVenueId())
		if err != nil {
			return nil, err
		}
		return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
	})
}

func (s *Server) GetTenantCourt(ctx context.Context, request *catalogV1.GetTenantCourtRequest) (*catalogV1.CourtResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.GetTenantCourtRequest) (*catalogV1.CourtResponse, error) {
		court, err := s.service.GetTenantCourt(ctx, actorID, request.GetCourtId())
		if err != nil {
			return nil, err
		}
		return &catalogV1.CourtResponse{Court: mapCourt(court)}, nil
	})
}
