package tenantgrpc

import (
	"context"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *Server) SetOrganizationStaff(
	ctx context.Context,
	req *identityV1.TenantSetOrganizationStaffRequest,
) (*identityV1.TenantStaffResponse, error) {
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
			req *identityV1.TenantSetOrganizationStaffRequest,
		) (*identityV1.TenantStaffResponse, error) {
			// Update organization staff.
			item, err := s.service.SetOrganizationStaff(ctx, actor, tenant.SetOrganizationStaffInput{
				OrganizationID: req.GetOrganizationId(),
				UserID:         req.GetUserId(),
				Role:           req.GetRole(),
				Status:         req.GetStatus(),
			})
			if err != nil {
				return nil, err
			}
			return organizationStaffResponse(item), nil
		},
	)
}

func (s *Server) SetVenueStaff(
	ctx context.Context,
	req *identityV1.TenantSetVenueStaffRequest,
) (*identityV1.TenantVenueStaffResponse, error) {
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
			req *identityV1.TenantSetVenueStaffRequest,
		) (*identityV1.TenantVenueStaffResponse, error) {
			// Update venue staff.
			item, err := s.service.SetVenueStaff(ctx, actor, tenant.SetVenueStaffInput{
				VenueID: req.GetVenueId(),
				UserID:  req.GetUserId(),
				Status:  req.GetStatus(),
			})
			if err != nil {
				return nil, err
			}
			return venueStaffResponse(item), nil
		},
	)
}
