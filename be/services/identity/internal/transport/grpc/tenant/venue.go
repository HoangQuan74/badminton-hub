package tenantgrpc

import (
	"context"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *Server) ListPublicVenues(
	ctx context.Context,
	req *identityV1.TenantListPublicVenuesRequest,
) (*identityV1.TenantVenuesResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		req,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			_ *identityV1.TenantListPublicVenuesRequest,
		) (*identityV1.TenantVenuesResponse, error) {
			// Load public venues.
			rows, err := s.service.ListPublicVenues(ctx)
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantVenuesResponse{Items: mapVenues(rows)}, nil
		},
	)
}

func (s *Server) GetPublicVenue(
	ctx context.Context,
	req *identityV1.TenantGetPublicVenueRequest,
) (*identityV1.TenantVenueResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		req,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			req *identityV1.TenantGetPublicVenueRequest,
		) (*identityV1.TenantVenueResponse, error) {
			// Load public venue.
			item, err := s.service.GetPublicVenue(ctx, req.GetSlug())
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
		},
	)
}

func (s *Server) GetPublicVenueByID(
	ctx context.Context,
	req *identityV1.TenantGetPublicVenueByIDRequest,
) (*identityV1.TenantVenueResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		req,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			req *identityV1.TenantGetPublicVenueByIDRequest,
		) (*identityV1.TenantVenueResponse, error) {
			// Load public venue by id.
			item, err := s.service.GetPublicVenueByID(ctx, req.GetVenueId())
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
		},
	)
}

func (s *Server) ListVenues(
	ctx context.Context,
	req *identityV1.TenantListVenuesRequest,
) (*identityV1.TenantVenuesResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(
		ctx,
		req,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			_ *identityV1.TenantListVenuesRequest,
		) (*identityV1.TenantVenuesResponse, error) {
			// Load venues.
			rows, err := s.service.ListVenues(ctx, actor)
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantVenuesResponse{Items: mapVenues(rows)}, nil
		},
	)
}

func (s *Server) GetVenue(
	ctx context.Context,
	req *identityV1.TenantGetVenueRequest,
) (*identityV1.TenantVenueResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(
		ctx,
		req,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			req *identityV1.TenantGetVenueRequest,
		) (*identityV1.TenantVenueResponse, error) {
			// Verify venue.
			item, err := s.service.AuthorizeVenue(
				ctx,
				actor,
				req.GetVenueId(),
				false,
			)
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
		},
	)
}

func (s *Server) AuthorizeVenue(
	ctx context.Context,
	req *identityV1.TenantAuthorizeVenueRequest,
) (*identityV1.TenantVenueResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(
		ctx,
		req,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			req *identityV1.TenantAuthorizeVenueRequest,
		) (*identityV1.TenantVenueResponse, error) {
			// Verify venue.
			item, err := s.service.AuthorizeVenue(
				ctx,
				actor,
				req.GetVenueId(),
				req.GetRequireOwner(),
			)
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
		},
	)
}

func (s *Server) CreateVenue(
	ctx context.Context,
	req *identityV1.TenantCreateVenueRequest,
) (*identityV1.TenantVenueResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(
		ctx,
		req,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			req *identityV1.TenantCreateVenueRequest,
		) (*identityV1.TenantVenueResponse, error) {
			// Create venue.
			item, err := s.service.CreateVenue(ctx, actor, tenant.CreateVenueInput{
				OrganizationID: req.GetOrganizationId(),
				Name:           req.GetName(),
				Slug:           req.GetSlug(),
				Address:        req.GetAddress(),
				Timezone:       req.GetTimezone(),
			})
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
		},
	)
}

func (s *Server) UpdateVenue(
	ctx context.Context,
	req *identityV1.TenantUpdateVenueRequest,
) (*identityV1.TenantVenueResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(
		ctx,
		req,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			req *identityV1.TenantUpdateVenueRequest,
		) (*identityV1.TenantVenueResponse, error) {
			// Update venue.
			item, err := s.service.UpdateVenue(ctx, actor, req.GetVenueId(), tenant.UpdateVenueInput{
				Name:     req.Name,
				Slug:     req.Slug,
				Address:  req.Address,
				Timezone: req.Timezone,
			})
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
		},
	)
}

func (s *Server) ArchiveVenue(
	ctx context.Context,
	req *identityV1.TenantArchiveVenueRequest,
) (*identityV1.TenantVenueResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}
	return grpcEndpoint.Run(
		ctx,
		req,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			req *identityV1.TenantArchiveVenueRequest,
		) (*identityV1.TenantVenueResponse, error) {
			// Archive venue.
			item, err := s.service.ArchiveVenue(
				ctx,
				actor,
				req.GetVenueId(),
			)
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantVenueResponse{Venue: venueResponse(item)}, nil
		},
	)
}
