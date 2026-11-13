package cataloggrpc

import (
	"context"

	catalogException "github.com/kainguyen/badminton-hub/be/services/catalog/internal/exception"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
)

func (s *CatalogServer) ListPublicVenueCourts(
	ctx context.Context,
	request *catalogV1.ListPublicVenueCourtsRequest,
) (*catalogV1.ListCourtsResponse, error) {
	courts, err := s.service.ListPublicVenueCourts(ctx, request.GetVenueId())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.ListCourtsResponse{Items: mapCourts(courts)}, nil
}

func (s *CatalogServer) ListTenantCourts(
	ctx context.Context,
	request *catalogV1.ListTenantCourtsRequest,
) (*catalogV1.ListCourtsResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}

	courts, err := s.service.ListTenantCourts(ctx, actorID, request.GetVenueId())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.ListCourtsResponse{Items: mapCourts(courts)}, nil
}

func (s *CatalogServer) CreateCourt(ctx context.Context, request *catalogV1.CreateCourtRequest) (*catalogV1.CourtResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}

	court, err := s.service.CreateCourt(ctx, actorID, request.GetVenueId(), request.GetName(), request.GetSlotIntervalMinutes())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.CourtResponse{Court: mapCourt(court)}, nil
}

func (s *CatalogServer) UpdateCourt(ctx context.Context, request *catalogV1.UpdateCourtRequest) (*catalogV1.CourtResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}

	court, err := s.service.UpdateCourt(
		ctx,
		actorID,
		request.GetCourtId(),
		request.GetName(),
		request.GetSlotIntervalMinutes(),
		request.GetStatus(),
	)
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.CourtResponse{Court: mapCourt(court)}, nil
}

func (s *CatalogServer) ArchiveCourt(ctx context.Context, request *catalogV1.ArchiveCourtRequest) (*catalogV1.ArchiveCourtResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}

	court, err := s.service.ArchiveCourt(ctx, actorID, request.GetCourtId())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.ArchiveCourtResponse{Court: mapCourt(court)}, nil
}

func (s *CatalogServer) GetTenantCourt(ctx context.Context, request *catalogV1.GetTenantCourtRequest) (*catalogV1.CourtResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}

	court, err := s.service.GetTenantCourt(ctx, actorID, request.GetCourtId())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.CourtResponse{Court: mapCourt(court)}, nil
}
