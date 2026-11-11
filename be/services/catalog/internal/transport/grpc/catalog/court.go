package cataloggrpc

import (
	"context"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
)

func (s *Server) ListPublicVenueCourts(
	ctx context.Context,
	request *catalogV1.ListPublicVenueCourtsRequest,
) (*catalogV1.ListCourtsResponse, error) {
	return grpcEndpoint.Run(
		ctx,
		request,
		errorMapper,
		nil,
		func(
			ctx context.Context,
			request *catalogV1.ListPublicVenueCourtsRequest,
		) (*catalogV1.ListCourtsResponse, error) {
			// Load public venue courts.
			courts, err := s.service.ListPublicVenueCourts(ctx, request.GetVenueId())
			if err != nil {
				return nil, err
			}
			return &catalogV1.ListCourtsResponse{Items: mapCourts(courts)}, nil
		},
	)
}

func (s *Server) ListTenantCourts(
	ctx context.Context,
	request *catalogV1.ListTenantCourtsRequest,
) (*catalogV1.ListCourtsResponse, error) {
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
			request *catalogV1.ListTenantCourtsRequest,
		) (*catalogV1.ListCourtsResponse, error) {
			// Load tenant courts.
			courts, err := s.service.ListTenantCourts(
				ctx,
				actorID,
				request.GetVenueId(),
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.ListCourtsResponse{Items: mapCourts(courts)}, nil
		},
	)
}

func (s *Server) CreateCourt(
	ctx context.Context,
	request *catalogV1.CreateCourtRequest,
) (*catalogV1.CourtResponse, error) {
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
			request *catalogV1.CreateCourtRequest,
		) (*catalogV1.CourtResponse, error) {
			// Create court.
			court, err := s.service.CreateCourt(
				ctx,
				actorID,
				request.GetVenueId(),
				request.GetName(),
				request.GetSlotIntervalMinutes(),
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.CourtResponse{Court: mapCourt(court)}, nil
		},
	)
}

func (s *Server) UpdateCourt(
	ctx context.Context,
	request *catalogV1.UpdateCourtRequest,
) (*catalogV1.CourtResponse, error) {
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
			request *catalogV1.UpdateCourtRequest,
		) (*catalogV1.CourtResponse, error) {
			// Update court.
			court, err := s.service.UpdateCourt(
				ctx,
				actorID,
				request.GetCourtId(),
				request.GetName(),
				request.GetSlotIntervalMinutes(),
				request.GetStatus(),
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.CourtResponse{Court: mapCourt(court)}, nil
		},
	)
}

func (s *Server) ArchiveCourt(
	ctx context.Context,
	request *catalogV1.ArchiveCourtRequest,
) (*catalogV1.ArchiveCourtResponse, error) {
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
			request *catalogV1.ArchiveCourtRequest,
		) (*catalogV1.ArchiveCourtResponse, error) {
			// Archive court.
			court, err := s.service.ArchiveCourt(
				ctx,
				actorID,
				request.GetCourtId(),
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.ArchiveCourtResponse{Court: mapCourt(court)}, nil
		},
	)
}

func (s *Server) GetTenantCourt(
	ctx context.Context,
	request *catalogV1.GetTenantCourtRequest,
) (*catalogV1.CourtResponse, error) {
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
			request *catalogV1.GetTenantCourtRequest,
		) (*catalogV1.CourtResponse, error) {
			// Load tenant court.
			court, err := s.service.GetTenantCourt(
				ctx,
				actorID,
				request.GetCourtId(),
			)
			if err != nil {
				return nil, err
			}
			return &catalogV1.CourtResponse{Court: mapCourt(court)}, nil
		},
	)
}
