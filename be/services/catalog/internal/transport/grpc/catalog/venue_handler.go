package cataloggrpc

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	catalogException "github.com/kainguyen/badminton-hub/be/services/catalog/internal/exception"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
)

func (s *CatalogServer) GetPublicVenue(ctx context.Context, request *catalogV1.GetPublicVenueRequest) (*catalogV1.VenueResponse, error) {
	venue, err := s.service.GetPublicVenue(ctx, request.GetSlug())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
}

func (s *CatalogServer) GetTenantVenue(ctx context.Context, request *catalogV1.GetTenantVenueRequest) (*catalogV1.VenueResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}

	venue, err := s.service.GetTenantVenue(ctx, actorID, request.GetVenueId())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
}

func (s *CatalogServer) CreateVenue(ctx context.Context, request *catalogV1.CreateVenueRequest) (*catalogV1.VenueResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	input := catalog.CreateVenueInput{
		OrganizationID: request.GetOrganizationId(),
		Name:           request.GetName(),
		Slug:           request.GetSlug(),
		Address:        request.GetAddress(),
		Timezone:       request.GetTimezone(),
	}

	venue, err := s.service.CreateVenue(ctx, actorID, input)
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
}

func (s *CatalogServer) UpdateVenue(ctx context.Context, request *catalogV1.UpdateVenueRequest) (*catalogV1.VenueResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	input := catalog.UpdateVenueInput{
		Name:     request.Name,
		Slug:     request.Slug,
		Address:  request.Address,
		Timezone: request.Timezone,
	}

	venue, err := s.service.UpdateVenue(ctx, actorID, request.GetVenueId(), input)
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
}

func (s *CatalogServer) ArchiveVenue(ctx context.Context, request *catalogV1.ArchiveVenueRequest) (*catalogV1.VenueResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}

	venue, err := s.service.ArchiveVenue(ctx, actorID, request.GetVenueId())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.VenueResponse{Venue: mapVenue(venue)}, nil
}

func (s *CatalogServer) ListPublicVenues(
	ctx context.Context,
	request *catalogV1.ListPublicVenuesRequest,
) (*catalogV1.ListVenuesResponse, error) {
	venues, err := s.service.ListPublicVenues(ctx)
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.ListVenuesResponse{Items: mapVenues(venues)}, nil
}

func (s *CatalogServer) ListTenantVenues(
	ctx context.Context,
	request *catalogV1.ListTenantVenuesRequest,
) (*catalogV1.ListVenuesResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}

	venues, err := s.service.ListTenantVenues(ctx, actorID)
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.ListVenuesResponse{Items: mapVenues(venues)}, nil
}
