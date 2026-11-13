package venuegrpc

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/venue"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *VenueServer) ListPublicVenues(
	ctx context.Context,
	req *identityV1.TenantListPublicVenuesRequest,
) (*identityV1.TenantVenuesResponse, error) {
	// Load public venues.
	rows, err := s.service.ListPublicVenues(ctx)
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantVenuesResponse{Items: mapVenues(rows)}, nil
}

func (s *VenueServer) GetPublicVenue(
	ctx context.Context,
	req *identityV1.TenantGetPublicVenueRequest,
) (*identityV1.TenantVenueResponse, error) {
	// Load public venue.
	item, err := s.service.GetPublicVenue(ctx, req.GetSlug())
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
}

func (s *VenueServer) GetPublicVenueByID(
	ctx context.Context,
	req *identityV1.TenantGetPublicVenueByIDRequest,
) (*identityV1.TenantVenueResponse, error) {
	// Load public venue by id.
	item, err := s.service.GetPublicVenueByID(ctx, req.GetVenueId())
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
}

func (s *VenueServer) ListVenues(ctx context.Context, req *identityV1.TenantListVenuesRequest) (*identityV1.TenantVenuesResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Load venues.
	rows, err := s.service.ListVenues(ctx, actor)
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantVenuesResponse{Items: mapVenues(rows)}, nil
}

func (s *VenueServer) GetVenue(ctx context.Context, req *identityV1.TenantGetVenueRequest) (*identityV1.TenantVenueResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Verify venue.
	item, err := s.service.AuthorizeVenue(ctx, actor, req.GetVenueId(), false)
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
}

func (s *VenueServer) AuthorizeVenue(
	ctx context.Context,
	req *identityV1.TenantAuthorizeVenueRequest,
) (*identityV1.TenantVenueResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Verify venue.
	item, err := s.service.AuthorizeVenue(ctx, actor, req.GetVenueId(), req.GetRequireOwner())
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
}

func (s *VenueServer) CreateVenue(ctx context.Context, req *identityV1.TenantCreateVenueRequest) (*identityV1.TenantVenueResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Create venue.
	input := venue.CreateVenueInput{
		OrganizationID: req.GetOrganizationId(),
		Name:           req.GetName(),
		Slug:           req.GetSlug(),
		Address:        req.GetAddress(),
		Timezone:       req.GetTimezone(),
	}

	item, err := s.service.CreateVenue(ctx, actor, input)
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
}

func (s *VenueServer) UpdateVenue(ctx context.Context, req *identityV1.TenantUpdateVenueRequest) (*identityV1.TenantVenueResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Update venue.
	input := venue.UpdateVenueInput{
		Name:     req.Name,
		Slug:     req.Slug,
		Address:  req.Address,
		Timezone: req.Timezone,
	}

	item, err := s.service.UpdateVenue(ctx, actor, req.GetVenueId(), input)
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
}

func (s *VenueServer) ArchiveVenue(ctx context.Context, req *identityV1.TenantArchiveVenueRequest) (*identityV1.TenantVenueResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Archive venue.
	item, err := s.service.ArchiveVenue(ctx, actor, req.GetVenueId())
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
}
