package tenantgrpc

import (
	"context"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *Server) ListOrganizations(
	ctx context.Context,
	req *identityV1.TenantListOrganizationsRequest,
) (*identityV1.TenantOrganizationsResponse, error) {
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
			_ *identityV1.TenantListOrganizationsRequest,
		) (*identityV1.TenantOrganizationsResponse, error) {
			// Load organizations.
			rows, err := s.service.ListOrganizations(ctx, actor)
			if err != nil {
				return nil, err
			}
			items := make([]*identityV1.TenantOrganization, 0, len(rows))
			for _, row := range rows {
				items = append(items, organizationResponse(row))
			}
			return &identityV1.TenantOrganizationsResponse{Items: items}, nil
		},
	)
}

func (s *Server) GetOrganization(
	ctx context.Context,
	req *identityV1.TenantGetOrganizationRequest,
) (*identityV1.TenantOrganizationResponse, error) {
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
			req *identityV1.TenantGetOrganizationRequest,
		) (*identityV1.TenantOrganizationResponse, error) {
			// Load organization.
			item, err := s.service.GetOrganization(
				ctx,
				actor,
				req.GetOrganizationId(),
			)
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantOrganizationResponse{Organization: organizationResponse(item)}, nil
		},
	)
}

func (s *Server) CreateOrganization(
	ctx context.Context,
	req *identityV1.TenantCreateOrganizationRequest,
) (*identityV1.TenantOrganizationResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), true)
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
			req *identityV1.TenantCreateOrganizationRequest,
		) (*identityV1.TenantOrganizationResponse, error) {
			// Create organization.
			item, err := s.service.CreateOrganization(
				ctx,
				actor,
				tenant.CreateOrganizationInput{Name: req.GetName(), Slug: req.GetSlug()},
			)
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantOrganizationResponse{Organization: organizationResponse(item)}, nil
		},
	)
}

func (s *Server) UpdateOrganization(
	ctx context.Context,
	req *identityV1.TenantUpdateOrganizationRequest,
) (*identityV1.TenantOrganizationResponse, error) {
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
			req *identityV1.TenantUpdateOrganizationRequest,
		) (*identityV1.TenantOrganizationResponse, error) {
			// Update organization.
			item, err := s.service.UpdateOrganization(
				ctx,
				actor,
				req.GetOrganizationId(),
				tenant.UpdateOrganizationInput{Name: req.Name, Slug: req.Slug},
			)
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantOrganizationResponse{Organization: organizationResponse(item)}, nil
		},
	)
}

func (s *Server) ArchiveOrganization(
	ctx context.Context,
	req *identityV1.TenantArchiveOrganizationRequest,
) (*identityV1.TenantOrganizationResponse, error) {
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
			req *identityV1.TenantArchiveOrganizationRequest,
		) (*identityV1.TenantOrganizationResponse, error) {
			// Archive organization.
			item, err := s.service.ArchiveOrganization(
				ctx,
				actor,
				req.GetOrganizationId(),
			)
			if err != nil {
				return nil, err
			}
			return &identityV1.TenantOrganizationResponse{Organization: organizationResponse(item)}, nil
		},
	)
}
