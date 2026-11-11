package cataloggrpc

import (
	"context"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
)

func (s *Server) GetPublicVenue(
	ctx context.Context,
	request *catalogV1.GetPublicVenueRequest,
) (*catalogV1.VenueResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *catalogV1.GetPublicVenueRequest,
		) (*catalogV1.VenueResponse, error) {
			// Load public venue.
			venue, err := s.service.GetPublicVenue(ctx, request.GetSlug())
			if err != nil {
				return nil, err
			}
			return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
		},
	)
}

func (s *Server) GetTenantVenue(
	ctx context.Context,
	request *catalogV1.GetTenantVenueRequest,
) (*catalogV1.VenueResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *catalogV1.GetTenantVenueRequest,
		) (*catalogV1.VenueResponse, error) {
			// Load tenant venue.
			venue, err := s.service.GetTenantVenue(
				ctx,
				actorID,
				request.GetVenueId(),
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
		},
	)
}

func (s *Server) CreateVenue(
	ctx context.Context,
	request *catalogV1.CreateVenueRequest,
) (*catalogV1.VenueResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *catalogV1.CreateVenueRequest,
		) (*catalogV1.VenueResponse, error) {
			input := catalog.CreateVenueInput{
				OrganizationID: request.GetOrganizationId(),
				Name:           request.GetName(),
				Slug:           request.GetSlug(),
				Address:        request.GetAddress(),
				Timezone:       request.GetTimezone(),
			}

			// Create venue.
			venue, err := s.service.CreateVenue(
				ctx,
				actorID,
				input,
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
		},
	)
}

func (s *Server) UpdateVenue(
	ctx context.Context,
	request *catalogV1.UpdateVenueRequest,
) (*catalogV1.VenueResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *catalogV1.UpdateVenueRequest,
		) (*catalogV1.VenueResponse, error) {
			input := catalog.UpdateVenueInput{
				Name:     request.Name,
				Slug:     request.Slug,
				Address:  request.Address,
				Timezone: request.Timezone,
			}

			// Update venue.
			venue, err := s.service.UpdateVenue(
				ctx,
				actorID,
				request.GetVenueId(),
				input,
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
		},
	)
}

func (s *Server) ArchiveVenue(
	ctx context.Context,
	request *catalogV1.ArchiveVenueRequest,
) (*catalogV1.VenueResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *catalogV1.ArchiveVenueRequest,
		) (*catalogV1.VenueResponse, error) {
			// Archive venue.
			venue, err := s.service.ArchiveVenue(
				ctx,
				actorID,
				request.GetVenueId(),
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
		},
	)
}

func (s *Server) ListPublicVenues(
	ctx context.Context,
	request *catalogV1.ListPublicVenuesRequest,
) (*catalogV1.ListVenuesResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *catalogV1.ListPublicVenuesRequest,
		) (*catalogV1.ListVenuesResponse, error) {
			// Load public venues.
			venues, err := s.service.ListPublicVenues(ctx)
			if err != nil {
				return nil, err
			}
			return &catalogV1.ListVenuesResponse{Items: mapVenues(venues)}, nil
		},
	)
}

func (s *Server) ListTenantVenues(
	ctx context.Context,
	request *catalogV1.ListTenantVenuesRequest,
) (*catalogV1.ListVenuesResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *catalogV1.ListTenantVenuesRequest,
		) (*catalogV1.ListVenuesResponse, error) {
			// Load tenant venues.
			venues, err := s.service.ListTenantVenues(ctx, actorID)
			if err != nil {
				return nil, err
			}
			return &catalogV1.ListVenuesResponse{Items: mapVenues(venues)}, nil
		},
	)
}
