package cataloggrpc

import (
	"context"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
)

func (s *Server) ListOrganizations(
	ctx context.Context,
	request *catalogV1.ListOrganizationsRequest,
) (*catalogV1.OrganizationListResponse, error) {
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
			_ *catalogV1.ListOrganizationsRequest,
		) (*catalogV1.OrganizationListResponse, error) {
			// Load organizations.
			items, err := s.service.ListOrganizations(ctx, actorID)
			if err != nil {
				return nil, err
			}
			results := make([]*catalogV1.Organization, 0, len(items))
			for _, item := range items {
				results = append(results, mapOrganization(item))
			}
			return &catalogV1.OrganizationListResponse{Items: results}, nil
		},
	)
}

func (s *Server) GetOrganization(
	ctx context.Context,
	request *catalogV1.GetOrganizationRequest,
) (*catalogV1.OrganizationResponse, error) {
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
			request *catalogV1.GetOrganizationRequest,
		) (*catalogV1.OrganizationResponse, error) {
			// Load organization.
			item, err := s.service.GetOrganization(
				ctx,
				actorID,
				request.GetOrganizationId(),
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.OrganizationResponse{Organization: mapOrganization(item)}, nil
		},
	)
}

func (s *Server) CreateOrganization(
	ctx context.Context,
	request *catalogV1.CreateOrganizationRequest,
) (*catalogV1.OrganizationResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantOwnerID(ctx)
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
			request *catalogV1.CreateOrganizationRequest,
		) (*catalogV1.OrganizationResponse, error) {
			// Create organization.
			item, err := s.service.CreateOrganization(
				ctx,
				actorID,
				catalog.CreateOrganizationInput{Name: request.GetName(), Slug: request.GetSlug()},
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.OrganizationResponse{Organization: mapOrganization(item)}, nil
		},
	)
}

func (s *Server) UpdateOrganization(
	ctx context.Context,
	request *catalogV1.UpdateOrganizationRequest,
) (*catalogV1.OrganizationResponse, error) {
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
			request *catalogV1.UpdateOrganizationRequest,
		) (*catalogV1.OrganizationResponse, error) {
			// Update organization.
			item, err := s.service.UpdateOrganization(
				ctx,
				actorID,
				request.GetOrganizationId(),
				catalog.UpdateOrganizationInput{Name: request.Name, Slug: request.Slug},
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.OrganizationResponse{Organization: mapOrganization(item)}, nil
		},
	)
}

func (s *Server) ArchiveOrganization(
	ctx context.Context,
	request *catalogV1.ArchiveOrganizationRequest,
) (*catalogV1.OrganizationResponse, error) {
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
			request *catalogV1.ArchiveOrganizationRequest,
		) (*catalogV1.OrganizationResponse, error) {
			// Archive organization.
			item, err := s.service.ArchiveOrganization(
				ctx,
				actorID,
				request.GetOrganizationId(),
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.OrganizationResponse{Organization: mapOrganization(item)}, nil
		},
	)
}
